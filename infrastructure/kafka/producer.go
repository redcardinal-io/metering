package kafka

import (
	"fmt"
	"strings"
	"sync"

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
	eventQueue   chan eventTask
	workerCount  int
	wg           sync.WaitGroup
	shutdownChan chan struct{}
}

type eventTask struct {
	topic string
	event *models.Event
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

	// Default to 5 workers, but could be configurable
	workerCount := 5
	if config.KafkaWorkerCount > 0 {
		workerCount = config.KafkaWorkerCount
	}

	// Buffer size for the event queue
	queueSize := 1000
	if config.KafkaQueueSize > 0 {
		queueSize = config.KafkaQueueSize
	}

	repo := &KafkaProducerRepository{
		logger:       logger,
		publisher:    publisher,
		eventQueue:   make(chan eventTask, queueSize),
		workerCount:  workerCount,
		shutdownChan: make(chan struct{}),
	}

	// Start background workers
	repo.startWorkers()

	return repo, nil
}

// startWorkers starts background goroutines to process events
func (k *KafkaProducerRepository) startWorkers() {
	for i := range k.workerCount {
		k.wg.Add(1)
		go k.worker(i)
	}
}

// worker processes events from the queue
func (k *KafkaProducerRepository) worker(id int) {
	defer k.wg.Done()
	k.logger.Info(fmt.Sprintf("Starting Kafka publisher worker %d", id))

	for {
		select {
		case task, ok := <-k.eventQueue:
			if !ok {
				// Channel closed, exit worker
				k.logger.Info(fmt.Sprintf("Worker %d shutting down: channel closed", id))
				return
			}
			// Process the event
			k.publishEventSync(task.topic, task.event)
		case <-k.shutdownChan:
			// Received shutdown signal
			k.logger.Info(fmt.Sprintf("Worker %d shutting down: shutdown signal received", id))
			return
		}
	}
}

// publishEventSync publishes a single event synchronously (called by workers)
func (k *KafkaProducerRepository) publishEventSync(topic string, event *models.Event) {
	// Serialize the event to JSON
	eventData, err := event.ToJSON()
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to marshal event: %v", err))
		return
	}

	// Create a new message with a unique ID
	msg := message.NewMessage(watermill.NewUUID(), eventData)

	// Add metadata if needed
	msg.Metadata.Set("content-type", "application/json")

	// Publish the message to the Kafka topic
	if err := k.publisher.Publish(topic, msg); err != nil {
		k.logger.Error(fmt.Sprintf("Failed to publish event to topic %s: %v", topic, err))
		return
	}

	k.logger.Debug(fmt.Sprintf("Successfully published event to topic: %s", topic))
}

// PublishEvent queues a single event for asynchronous publishing
func (k *KafkaProducerRepository) PublishEvent(topic string, event *models.Event) error {
	select {
	case k.eventQueue <- eventTask{topic: topic, event: event}:
		// Successfully queued
		return nil
	default:
		// Queue is full
		return fmt.Errorf("event queue is full, cannot publish event")
	}
}

// PublishEvents queues multiple events for asynchronous publishing
func (k *KafkaProducerRepository) PublishEvents(topic string, eventBatch *models.EventBatch) error {
	k.logger.Info(fmt.Sprintf("Queueing %d events for publishing to topic: %s", len(eventBatch.Events), topic))

	for _, event := range eventBatch.Events {
		if err := k.PublishEvent(topic, event); err != nil {
			return fmt.Errorf("failed to queue event: %w", err)
		}
	}

	k.logger.Info(fmt.Sprintf("Successfully queued %d events for topic: %s", len(eventBatch.Events), topic))
	return nil
}

// Close stops all workers and closes the Kafka publisher
func (k *KafkaProducerRepository) Close() error {
	k.logger.Info("Closing Kafka producer repository")

	// Signal all workers to stop
	close(k.shutdownChan)

	// Wait for all workers to finish
	k.wg.Wait()

	// Close the event queue
	close(k.eventQueue)

	// Close the publisher
	return k.publisher.Close()
}
