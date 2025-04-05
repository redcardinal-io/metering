package kafka

import (
	"encoding/json"
	"fmt"
	"strings"

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
	logger    *logger.Logger
	publisher *kafka.Publisher
}

// NewKafkaProducerRepository creates a new instance of KafkaProducerRepository
func NewKafkaProducerRepository(logger *logger.Logger, config config.KafkaConfig) (repositories.ProducerRepository, error) {
	watermillLogger := watermill.NewStdLogger(false, false)

	// Parse brokers from bootstrap servers string
	brokers := strings.Split(config.KafkaBootstrapServers, ",")

	// Create publisher config
	publisherConfig := kafka.PublisherConfig{
		Brokers:   brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}

	// Only create a custom sarama config if we need SASL or other custom settings
	if config.KafkaUsername != "" && config.KafkaPassword != "" {
		// Configure sarama directly
		saramaConfig := sarama.NewConfig()

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

		// Set producer settings
		saramaConfig.Producer.Return.Successes = true
		saramaConfig.Producer.Return.Errors = true

		// Assign the sarama config to the publisher
		publisherConfig.OverwriteSaramaConfig = saramaConfig
	}

	publisher, err := kafka.NewPublisher(publisherConfig, watermillLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka publisher: %w", err)
	}

	return &KafkaProducerRepository{
		logger:    logger,
		publisher: publisher,
	}, nil
}

// PublishEvents publishes events to the specified Kafka topic
func (k *KafkaProducerRepository) PublishEvents(topic string, eventBatch *models.EventBatch) error {
	k.logger.Info(fmt.Sprintf("Publishing event batch to topic: %s", topic))

	// Serialize the event batch to JSON
	eventData, err := json.Marshal(eventBatch)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to marshal event batch: %v", err))
		return fmt.Errorf("failed to marshal event batch: %w", err)
	}

	// Create a new message with a unique ID
	msg := message.NewMessage(watermill.NewUUID(), eventData)

	// Add metadata if needed
	msg.Metadata.Set("content-type", "application/json")

	// Publish the message to the Kafka topic
	if err := k.publisher.Publish(topic, msg); err != nil {
		k.logger.Error(fmt.Sprintf("Failed to publish events to topic %s: %v", topic, err))
		return fmt.Errorf("failed to publish events to topic %s: %w", topic, err)
	}

	k.logger.Info(fmt.Sprintf("Successfully published event batch to topic: %s", topic))
	return nil
}

// Close closes the Kafka publisher
func (k *KafkaProducerRepository) Close() error {
	return k.publisher.Close()
}
