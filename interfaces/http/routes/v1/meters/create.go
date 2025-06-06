package meters

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
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

// @Summary Create a new meter
// @Description Create a new meter for the tenant
// @Tags meters
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param meter body createMeterRequest true "Meter creation data"
// @Success 201 {object} models.HttpResponse[models.Meter] "Meter created successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/meters [post]
func (h *httpHandler) create(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	var req createMeterRequest
	if err := ctx.BodyParser(&req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// validate the request body
	if err := h.validator.Struct(req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid request body")
		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	valueProperty := req.ValueProperty
	if req.Aggregation == string(models.AggregationCount) {
		valueProperty = ""
	} else if valueProperty == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(errors.New("value_property is required"), domainerrors.EINVALID, "value_property is required")
		h.logger.Error("value_property is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenant_slug)

	meter, err := h.meterSvc.CreateMeter(c, models.CreateMeterInput{
		Name:          req.Name,
		MeterSlug:     req.Slug,
		EventType:     req.EventType,
		Description:   req.Description,
		ValueProperty: valueProperty,
		Properties:    req.Properties,
		Aggregation:   models.AggregationEnum(req.Aggregation),
		Populate:      req.Populate,
		CreatedBy:     req.CreatedBy,
	})
	if err != nil {
		h.logger.Error("failed to create meter", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(meter, "meter created successfully", fiber.StatusCreated))
}
