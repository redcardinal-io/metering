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

func newHTTPHandler(logger *logger.Logger) *httpHandler {
	return &httpHandler{
		logger: logger,
	}
}

func RegisterRoutes(app *fiber.App, logger *logger.Logger) {
	handler := newHTTPHandler(logger)
	app.Get("/health", handler.healthCheck)
}
