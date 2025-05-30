package plans

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

// @Summary Get plan details
// @Description Get detailed information about a plan by ID or slug
// @Tags plans
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param idOrSlug path string true "Plan ID or slug"
// @Success 200 {object} models.HttpResponse[models.Plan] "Plan retrieved successfully"
// @Failure 400 {object} domainerrors.ErrorResponse  "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Plan not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{idOrSlug} [get]
func (h *httpHandler) details(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan ID is required")
		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	plan, err := h.planSvc.GetPlanByIDorSlug(c, idOrSlug)
	if err != nil {
		h.logger.Error("failed to get plan", zap.String("idOrSlug", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(plan, "plan retrieved successfully", fiber.StatusOK))
}
