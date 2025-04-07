package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type KafkaProducerRepository struct {
	logger       *logger.Logger
	publisher    *kafka.Publisher
	maxRetries   int
	retryBackoff time.Duration
}

// NewKafkaProducerRepository creates a new instance of KafkaProducerRepository
func NewKafkaProducerRepository(logger *logger.Logger, config config.KafkaConfig) (repositories.ProducerRepository, error) {
	watermillLogger := watermill.NewStdLogger(false, false)
	// Parse brokers from bootstrap servers string
	brokers := strings.Split(config.KafkaBootstrapServers, ",")

	// Create custom marshaler that properly handles message keys
	marshaler := &CustomMarshaler{}

	// Create publisher config
	publisherConfig := kafka.PublisherConfig{
		Brokers:   brokers,
		Marshaler: marshaler,
	}

	// Configure Sarama
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Partitioner = sarama.NewHashPartitioner

	// Configure SASL if credentials are provided
	if config.KafkaUsername != "" && config.KafkaPassword != "" {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.User = config.KafkaUsername
		saramaConfig.Net.SASL.Password = config.KafkaPassword

		// Set SASL mechanism if provided
		if config.KafkaSaslMechanisms != "" {
			saramaConfig.Net.SASL.Mechanism = sarama.SASLMechanism(config.KafkaSaslMechanisms)
		}

		// Set security protocol if provided
		if config.KafkaSecurityProtocol != "" {
			if config.KafkaSecurityProtocol == "SASL_SSL" {
				saramaConfig.Net.TLS.Enable = true
			} else if config.KafkaSecurityProtocol == "PLAINTEXT" {
				saramaConfig.Net.TLS.Enable = false
			} else {
				return nil, fmt.Errorf("unsupported security protocol: %s", config.KafkaSecurityProtocol)
			}
		}
	}

	publisherConfig.OverwriteSaramaConfig = saramaConfig

	publisher, err := kafka.NewPublisher(publisherConfig, watermillLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka publisher: %w", err)
	}

	// Default retry values if not configured
	maxRetries := 3
	if config.KafkaMaxRetries > 0 {
		maxRetries = config.KafkaMaxRetries
	}

	retryBackoff := 100 * time.Millisecond
	if config.KafkaRetryBackoffMs > 0 {
		retryBackoff = time.Duration(config.KafkaRetryBackoffMs) * time.Millisecond
	}

	repo := &KafkaProducerRepository{
		logger:       logger,
		publisher:    publisher,
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
	}

	return repo, nil
}

// publishEventSync publishes a single event synchronously
func (k *KafkaProducerRepository) publishEventSync(topic string, event *models.Event) error {
	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	partitionKey := fmt.Sprintf("%s-%s", event.Organization, event.User)

	msg := message.NewMessage(watermill.NewUUID(), eventData)
	msg.Metadata.Set("key", partitionKey)
	msg.Metadata.Set("content-type", "application/json")

	k.logger.Debug(fmt.Sprintf("Publishing message with partition key: %s", partitionKey))

	if err := k.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish event to topic %s: %w", topic, err)
	}

	k.logger.Debug(fmt.Sprintf("Successfully published event to topic: %s", topic))
	return nil
}

// PublishEvent publishes a single event with retries
func (k *KafkaProducerRepository) PublishEvent(topic string, event *models.Event) error {
	var err error

	for attempt := 0; attempt <= k.maxRetries; attempt++ {
		if attempt > 0 {
			// Apply exponential backoff for retries
			backoff := k.retryBackoff * time.Duration(1<<attempt) // 1<<attempt is equivalent to 2^attempt
			k.logger.Warn(fmt.Sprintf("Retrying event publish (attempt %d/%d) after %v due to: %v",
				attempt, k.maxRetries, backoff, err))
			time.Sleep(backoff)
		}

		err = k.publishEventSync(topic, event)
		if err == nil {
			return nil
		}
	}

	// If we get here, all attempts failed
	k.logger.Error(fmt.Sprintf("CRITICAL: Failed to publish event after %d attempts - EVENT LOST: %v",
		k.maxRetries+1, err))
	return err
}

// PublishEvents publishes multiple events with individual retries
func (k *KafkaProducerRepository) PublishEvents(topic string, eventBatch *models.EventBatch) error {
	k.logger.Info(fmt.Sprintf("Publishing %d events to topic: %s", len(eventBatch.Events), topic))

	var errs []error
	for _, event := range eventBatch.Events {
		if err := k.PublishEvent(topic, event); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to publish %d/%d events", len(errs), len(eventBatch.Events))
	}

	k.logger.Info(fmt.Sprintf("Successfully published %d events to topic: %s", len(eventBatch.Events), topic))
	return nil
}

// Close closes the Kafka publisher
func (k *KafkaProducerRepository) Close() error {
	k.logger.Info("Closing Kafka producer repository")
	return k.publisher.Close()
}
