package kafka

import (
	"fmt"
	"strings"
	"sync"
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
	eventQueue   chan eventTask
	workerCount  int
	wg           sync.WaitGroup
	shutdownChan chan struct{}

	// Retry configuration
	maxRetries   int
	retryBackoff time.Duration
}

type eventTask struct {
	topic      string
	event      *models.Event
	retryCount int
	createTime time.Time
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
		saramaConfig.Producer.Partitioner = sarama.NewHashPartitioner

		// Assign the sarama config to the publisher
		publisherConfig.OverwriteSaramaConfig = saramaConfig
	} else {
		saramaConfig := sarama.NewConfig()
		saramaConfig.Producer.Return.Successes = true
		saramaConfig.Producer.Return.Errors = true
		saramaConfig.Producer.Partitioner = sarama.NewHashPartitioner
		publisherConfig.OverwriteSaramaConfig = saramaConfig
	}

	publisher, err := kafka.NewPublisher(publisherConfig, watermillLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka publisher: %w", err)
	}

	workerCount := config.KafkaWorkerCount
	queueSize := config.KafkaQueueSize

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
		eventQueue:   make(chan eventTask, queueSize),
		workerCount:  workerCount,
		shutdownChan: make(chan struct{}),
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
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

func (k *KafkaProducerRepository) worker(id int) {
	defer k.wg.Done()
	k.logger.Info(fmt.Sprintf("Starting Kafka publisher worker %d", id))

	for {
		select {
		case task, ok := <-k.eventQueue:
			if !ok {
				k.logger.Info(fmt.Sprintf("Worker %d shutting down: channel closed", id))
				return
			}
			if err := k.publishEventSync(task.topic, task.event); err != nil {
				k.handlePublishError(task, err)
			}
		case <-k.shutdownChan:
			k.logger.Info(fmt.Sprintf("Worker %d shutting down: shutdown signal received", id))
			return
		}
	}
}

// handlePublishError attempts to retry publishing events based on retry policy
func (k *KafkaProducerRepository) handlePublishError(task eventTask, err error) {
	task.retryCount++

	if task.retryCount <= k.maxRetries {
		backoff := k.retryBackoff

		k.logger.Warn(fmt.Sprintf("Retrying event publish (attempt %d/%d) after %v: %v",
			task.retryCount, k.maxRetries, backoff, err))

		go func(t eventTask, b time.Duration) {
			time.Sleep(b)

			select {
			case k.eventQueue <- t:
				k.logger.Debug(fmt.Sprintf("Successfully requeued event after backoff (attempt %d)", t.retryCount))
			default:
				k.logger.Error(fmt.Sprintf("CRITICAL: Failed to requeue event after %d attempts, queue still full - EVENT LOST", t.retryCount))
			}
		}(task, backoff)
	} else {
		elapsed := time.Since(task.createTime)
		k.logger.Error(fmt.Sprintf("CRITICAL: Failed to publish event after %d attempts over %v - EVENT LOST: %v",
			task.retryCount, elapsed, err))
	}
}

// publishEventSync publishes a single event synchronously (called by workers)
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

// PublishEvent queues a single event for asynchronous publishing
func (k *KafkaProducerRepository) PublishEvent(topic string, event *models.Event) error {
	task := eventTask{
		topic:      topic,
		event:      event,
		retryCount: 0,
		createTime: time.Now(),
	}

	select {
	case k.eventQueue <- task:
		// Successfully queued
		return nil
	case <-time.After(100 * time.Millisecond):
		k.logger.Warn("Event queue is full, attempting to retry with backoff")

		go k.retryQueueWithBackoff(task, 1)

		return nil
	}
}

// retryQueueWithBackoff attempts to queue an event with exponential backoff
func (k *KafkaProducerRepository) retryQueueWithBackoff(task eventTask, attempt int) {
	maxQueueAttempts := 5

	if attempt > maxQueueAttempts {
		k.logger.Error(fmt.Sprintf("CRITICAL: Failed to queue event after %d attempts - EVENT LOST", attempt-1))
		return
	}

	backoff := time.Duration(50*attempt*attempt) * time.Millisecond
	time.Sleep(backoff)

	select {
	case k.eventQueue <- task:
		k.logger.Info(fmt.Sprintf("Successfully queued event after %d attempts", attempt))
		return
	default:
		k.logger.Warn(fmt.Sprintf("Queue still full, retry attempt %d/%d after %v",
			attempt, maxQueueAttempts, backoff))
		k.retryQueueWithBackoff(task, attempt+1)
	}
}

// PublishEvents queues multiple events for asynchronous publishing
func (k *KafkaProducerRepository) PublishEvents(topic string, eventBatch *models.EventBatch) error {
	k.logger.Info(fmt.Sprintf("Queueing %d events for publishing to topic: %s", len(eventBatch.Events), topic))

	var failedCount int
	for _, event := range eventBatch.Events {
		if err := k.PublishEvent(topic, event); err != nil {
			failedCount++
			k.logger.Error(fmt.Sprintf("Failed to queue event: %v", err))
		}
	}

	if failedCount > 0 {
		k.logger.Warn(fmt.Sprintf("%d/%d events failed to queue initially, retry logic engaged",
			failedCount, len(eventBatch.Events)))
	}

	k.logger.Info(fmt.Sprintf("Finished queueing events for topic: %s", topic))
	return nil // We always return success since retry is handled asynchronously
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
