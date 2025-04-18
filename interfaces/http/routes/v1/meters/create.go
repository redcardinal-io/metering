package meters

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/interfaces/http/routes/constants"
)

type createMeterRequest struct {
	Name          string   `json:"name" validate:"required"`
	Slug          string   `json:"slug" validate:"required"`
	EventType     string   `json:"event_type" validate:"required"`
	Description   string   `json:"description,omitempty"`
	ValueProperty string   `json:"value_property,omitempty"`
	Properties    []string `json:"properties" validate:"required,min=1"`
	Aggregation   string   `json:"aggregation" validate:"required,oneof=count sum avg unique_count min max"`
	CreatedBy     string   `json:"created_by" validate:"required"`
	Populate      bool     `json:"populate" validate:"required"`
}

func (h *httpHandler) create(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	if tenant_slug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant slug is required",
		})
	}

	var req createMeterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	meter, err := h.meterSvc.CreateMeter(ctx.UserContext(), models.CreateMeterInput{
		Name:          req.Name,
		MeterSlug:     req.Slug,
		EventType:     req.EventType,
		Description:   req.Description,
		ValueProperty: req.ValueProperty,
		Properties:    req.Properties,
		Aggregation:   models.AggregationEnum(req.Aggregation),
		CreatedBy:     req.CreatedBy,
		Populate:      req.Populate,
		TenantSlug:    tenant_slug,
	})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create meter",
		})
	}

	return ctx.
		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(meter, "meter created successfully", fiber.StatusCreated))
}
