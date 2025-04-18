package kafka

import (
	"errors"
	"strings"

	"github.com/Shopify/sarama"

	domainerrors "github.com/redcardinal-io/metering/domain/errors"
)

// MapError translates Kafka errors to domain errors
func MapError(err error, op string) error {
	if err == nil {
		return nil
	}

	// Check for specific Sarama errors using errors.Is
	// Topic errors
	if errors.Is(err, sarama.ErrTopicAuthorizationFailed) {
		return domainerrors.New(
			err,
			domainerrors.EFORBIDDEN,
			"Not authorized to access topic",
			domainerrors.WithOperation(op),
		)
	}

	if errors.Is(err, sarama.ErrUnknownTopicOrPartition) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"Topic does not exist or partition unavailable",
			domainerrors.WithOperation(op),
		)
	}

	if errors.Is(err, sarama.ErrInvalidTopic) || errors.Is(err, sarama.ErrInvalidPartitions) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"Invalid topic configuration",
			domainerrors.WithOperation(op),
		)
	}

	// Message errors
	if errors.Is(err, sarama.ErrMessageTooLarge) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"Message is too large",
			domainerrors.WithOperation(op),
		)
	}

	if errors.Is(err, sarama.ErrInvalidMessage) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"Invalid message format",
			domainerrors.WithOperation(op),
		)
	}

	// Authentication errors
	if errors.Is(err, sarama.ErrSASLAuthenticationFailed) {
		return domainerrors.New(
			err,
			domainerrors.EUNAUTHORIZED,
			"Failed to authenticate with Kafka",
			domainerrors.WithOperation(op),
		)
	}

	// Broker errors
	if errors.Is(err, sarama.ErrBrokerNotAvailable) || errors.Is(err, sarama.ErrClosedClient) {
		return domainerrors.New(
			err,
			domainerrors.EUNAVAILABLE,
			"Kafka broker unavailable",
			domainerrors.WithOperation(op),
		)
	}

	if errors.Is(err, sarama.ErrLeaderNotAvailable) || errors.Is(err, sarama.ErrReplicaNotAvailable) {
		return domainerrors.New(
			err,
			domainerrors.EUNAVAILABLE,
			"Kafka topic leader or replica unavailable",
			domainerrors.WithOperation(op),
		)
	}

	// Network errors
	if errors.Is(err, sarama.ErrNetworkException) {
		return domainerrors.New(
			err,
			domainerrors.EUNAVAILABLE,
			"Kafka network error",
			domainerrors.WithOperation(op),
		)
	}

	// Configuration errors
	if errors.Is(err, sarama.ErrInvalidConfig) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"Invalid Kafka configuration",
			domainerrors.WithOperation(op),
		)
	}

	// Timeout errors
	if errors.Is(err, sarama.ErrRequestTimedOut) {
		return domainerrors.New(
			err,
			domainerrors.ETIMEOUT,
			"Kafka operation timed out",
			domainerrors.WithOperation(op),
		)
	}

	// If it's not a specific Sarama error, check for common patterns in error strings
	// This is a fallback for errors that don't use the errors.Is mechanism
	errMsg := strings.ToLower(err.Error())

	if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "broker not available") {
		return domainerrors.New(
			err,
			domainerrors.EUNAVAILABLE,
			"Kafka service unavailable",
			domainerrors.WithOperation(op),
		)
	}

	if strings.Contains(errMsg, "authentication failed") || strings.Contains(errMsg, "sasl authentication failed") {
		return domainerrors.New(
			err,
			domainerrors.EUNAUTHORIZED,
			"Failed to authenticate with Kafka",
			domainerrors.WithOperation(op),
		)
	}

	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
		return domainerrors.New(
			err,
			domainerrors.ETIMEOUT,
			"Kafka operation timed out",
			domainerrors.WithOperation(op),
		)
	}

	// Fallback for unhandled Kafka errors
	return domainerrors.New(
		err,
		domainerrors.EMESSAGEBROKER,
		"Kafka error",
		domainerrors.WithOperation(op),
	)
}
