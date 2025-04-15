package events

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/domain/models"
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
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	if len(body.Events) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Empty event batch",
			"code":    fiber.StatusBadRequest,
			"message": "Event batch cannot be empty",
		})
	}

	events := models.EventBatch{}
	events.Events = make([]*models.Event, 0, len(body.Events))

	for _, event := range body.Events {
		if event.ID == "" {
			id, _ := uuid.NewV7()
			event.ID = id.String()
		}
		if event.Timestamp == "" {
			event.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		} else {
			_, err := time.Parse("2006-01-02 15:04:05", event.Timestamp)
			if err != nil {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Invalid timestamp format",
					"code":    fiber.StatusBadRequest,
					"message": err.Error(),
				})
			}
		}
		properties, err := json.Marshal(event.Properties)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid properties format",
				"code":    fiber.StatusBadRequest,
				"message": err.Error(),
			})
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to publish event",
			"code":    fiber.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusNoContent).JSON(nil)
}
