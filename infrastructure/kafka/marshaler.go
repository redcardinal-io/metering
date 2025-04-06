package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

// Create a custom marshaler that properly handles message keys
type CustomMarshaler struct {
	kafka.DefaultMarshaler
}

// Marshal properly sets the Kafka message key from metadata
func (m *CustomMarshaler) Marshal(topic string, msg *message.Message) (*sarama.ProducerMessage, error) {
	producerMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg.Payload),
	}

	// Extract key from metadata and set it as the actual Kafka message key
	if key, ok := msg.Metadata["key"]; ok {
		producerMsg.Key = sarama.StringEncoder(key)
	}

	// Copy other metadata to headers if needed
	for k, v := range msg.Metadata {
		if k == "key" {
			continue // Already handled
		}
		producerMsg.Headers = append(producerMsg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	return producerMsg, nil
}
