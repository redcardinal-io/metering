package plans

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"go.uber.org/zap"
)

func (h *httpHandler) list(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	planType := ctx.Params("type")

	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)
	var plans *pagination.PaginationView[models.Plan]
	var err error

	if planType == "" {
		// Call service to list all plans
		plans, err = h.planSvc.ListPlans(c, paginationInput)
	} else {

		if !models.ValidatePlanType(planType) {
			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid plan type")
			h.logger.Error("invalid plan type", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}

		parsedPlanType := models.PlanTypeEnum(planType)
		// Call service to list plans by type
		plans, err = h.planSvc.ListPlansByType(c, parsedPlanType, paginationInput)
	}

	if err != nil {
		h.logger.Error("failed to list plans", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}
	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(plans, "plans retrieved successfully", fiber.StatusOK))
}
