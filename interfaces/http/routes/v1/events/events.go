package events

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/domain/models"
)

func (h *httpHandler) publishEvent(ctx *fiber.Ctx) error {
	var events models.EventBatch

	if err := ctx.BodyParser(&events); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
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
