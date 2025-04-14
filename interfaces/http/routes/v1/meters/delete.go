package meters

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/domain/models"
)

func (h *httpHandler) delete(ctx *fiber.Ctx) error {

	var req deleteMeterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	meter, err := h.meterSvc.DeleteMeter(ctx.UserContext(), ctx.Get("organization"), req.Slug)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete meter",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"meter": meter,
	})
}
