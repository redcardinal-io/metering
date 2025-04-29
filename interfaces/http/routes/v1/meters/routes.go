package meters

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type httpHandler struct {
	logger    *logger.Logger
	meterSvc  *services.MeterService
	validator *validator.Validate
}

// NewHTTPHandler creates and returns a new httpHandler for meter-related HTTP endpoints.
func NewHTTPHandler(logger *logger.Logger, meterSvc *services.MeterService) *httpHandler {
	validator := validator.New()
	return &httpHandler{
		logger:    logger,
		meterSvc:  meterSvc,
		validator: validator,
	}
}

func (h *httpHandler) RegisterRoutes(r fiber.Router) {
	// Group all meter routes
	meters := r.Group("/meters")

	// Meter collection routes
	meters.Post("/", h.create)
	meters.Post("/query", h.query)
	meters.Get("/", h.list)

	// Single meter routes with idOrSlug parameter
	meters.Get("/:idOrSlug", h.getByIDorSlug)
	meters.Put("/:idOrSlug", h.updateByIDorSlug)
	meters.Delete("/:idOrSlug", h.deleteByIDorSlug)
}
