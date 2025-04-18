package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/interfaces/http/routes/constants"
	"go.uber.org/zap"
)

type event struct {
	// The event ID.
	ID string `json:"id"`
	// The event type.
	Type string `json:"type"`
	// The event source.
	Source string `json:"source"`
	// ID of the organization that user belongs to.
	Organization string `json:"organization"`
	// The ID of the user that owns the event.
	User string `json:"user"`
	// The event time.
	Timestamp string `json:"timestamp"`
	// The event data as a JSON string.
	Properties map[string]any `json:"properties"`
}

type publisEventRequestBody struct {
	Events []event `json:"events"`
}

func (h *httpHandler) publishEvent(ctx *fiber.Ctx) error {
	var body publisEventRequestBody

	if err := ctx.BodyParser(&body); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
		h.logger.Error("failed to parse request body", zap.Any("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if len(body.Events) == 0 {
		errResp := domainerrors.NewErrorResponseWithOpts(fmt.Errorf("cannot process empty event batch"), domainerrors.EINVALID, "empty event batch")
		h.logger.Error("empty event batch", zap.Any("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	events := models.EventBatch{}
	events.Events = make([]*models.Event, 0, len(body.Events))

	for _, event := range body.Events {
		if event.ID == "" {
			id, _ := uuid.NewV7()
			event.ID = id.String()
		}
		if event.Timestamp == "" {
			event.Timestamp = time.Now().Format(constants.TimeFormat)
		} else {
			_, err := time.Parse(constants.TimeFormat, event.Timestamp)
			if err != nil {
				errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid timestamp format")
				return ctx.Status(errResp.Status).JSON(errResp.ToJson())
			}
		}
		properties, err := json.Marshal(event.Properties)
		if err != nil {
			h.logger.Error("failed to parse properties", zap.Error(err))
			errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse properties")
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
		events.Events = append(events.Events, &models.Event{
			ID:           event.ID,
			Type:         event.Type,
			Source:       event.Source,
			Organization: event.Organization,
			User:         event.User,
			Timestamp:    event.Timestamp,
			Properties:   string(properties),
		})
	}

	err := h.producer.PublishEvents(h.publishTopic, &events)
	if err != nil {
		h.logger.Error("failed to publish events", zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusNoContent).JSON(models.NewHttpResponse[any](nil, "events published successfully", fiber.StatusNoContent))
}
