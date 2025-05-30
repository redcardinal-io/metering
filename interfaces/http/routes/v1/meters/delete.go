package meters

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

// @Summary Delete a meter
// @Description Delete a meter by ID or slug
// @Tags meters
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param idOrSlug path string true "Meter ID or slug"
// @Success 204 "Meter deleted successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Meter not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/meters/{idOrSlug} [delete]
func (h *httpHandler) deleteByIDorSlug(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "meter ID or slug is required")
		h.logger.Error("meter ID or slug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	err := h.meterSvc.DeleteMeter(c, idOrSlug)
	if err != nil {
		h.logger.Error("failed to delete meter", zap.String("idOrSlug", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		SendStatus(fiber.StatusNoContent)
}
