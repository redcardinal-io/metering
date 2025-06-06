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

type queryPlanParams struct {
	Page     int    `query:"page" validate:"omitempty,min=1"`
	Limit    int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Type     string `query:"type" validate:"omitempty,oneof=standard custom"`
	Archived bool   `query:"archived" validate:"omitempty"`
}

// @Summary List all plans
// @Description Get a paginated list of all plans for the tenant
// @Tags plans
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} models.HttpResponse[pagination.PaginationView[models.Plan]] "Plans retrieved successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans [get]
func (h *httpHandler) list(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)

	// Parse and validate query parameters
	params := new(queryPlanParams)
	if err := ctx.QueryParser(params); err != nil {
		h.logger.Error("failed to parse query parameters", zap.Error(err))
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse query parameters")
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Validate the parsed parameters
	if err := h.validator.Struct(params); err != nil {
		h.logger.Error("invalid query parameters", zap.Error(err))
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid query parameters")
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	plans, err := h.planSvc.ListPlans(c, paginationInput)
	if err != nil {
		h.logger.Error("failed to list plans", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}
	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(plans, "plans retrieved successfully", fiber.StatusOK))
}
