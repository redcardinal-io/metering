package features

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type httpHandler struct {
	logger     *logger.Logger
	featureSvc *services.PlanManagementService
	validator  *validator.Validate
}

// NewHTTPHandler creates a new httpHandler for feature-related HTTP routes with logging, feature service, and validation capabilities.
func NewHTTPHandler(logger *logger.Logger, featureSvc *services.PlanManagementService) *httpHandler {
	validator := validator.New()
	return &httpHandler{logger, featureSvc, validator}
}

func (h *httpHandler) RegisterRoutes(r fiber.Router) {
	features := r.Group("/features")

	features.Post("/", h.create)
	features.Get("/", h.list)
	features.Get("/:idOrSlug", h.getByIDorSlug)
	features.Put("/:idOrSlug", h.updateByIDorSlug)
	features.Delete("/:idOrSlug", h.deleteByIDorSlug)
}
