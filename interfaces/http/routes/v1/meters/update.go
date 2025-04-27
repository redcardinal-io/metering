package meters

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type updateMeterRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,min=3,max=255"`
}

func (h *httpHandler) updateByIDorSlug(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "meter ID or slug is required")
		h.logger.Error("meter ID or slug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req updateMeterRequest
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
	if req.Name == "" && req.Description == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "at least one field (name or description) is required")
		h.logger.Error("at least one field (name or description) is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	meter, err := h.meterSvc.UpdateMeter(c, idOrSlug, models.UpdateMeterInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.logger.Error("failed to update meter", zap.String("idOrSlug", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(meter, "meter updated successfully", fiber.StatusOK))
}
