package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

func (h *httpHandler) healthCheck(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

type httpHandler struct {
	logger *logger.Logger
}

func NewHTTPHandler(logger *logger.Logger) *httpHandler {
	return &httpHandler{
		logger: logger,
	}
}

func (h *httpHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.healthCheck)
}
