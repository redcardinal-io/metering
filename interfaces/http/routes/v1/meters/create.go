package meters

import (
	"context"

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
	})
	if err != nil {
		h.logger.Error("failed to create meter", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(meter, "meter created successfully", fiber.StatusCreated))
}
