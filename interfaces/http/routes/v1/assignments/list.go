package assignments

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
	orgId := ctx.Query("orgId")
	userId := ctx.Query("userId")

	if orgId != "" && userId != "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "provide only one profile param")
		h.logger.Error("provide only one profile param", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}
	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	var planassignments *pagination.PaginationView[models.PlanAssignment]

	if orgId != "" {
		// Call service to list org assignments
		h.logger.Info("org", zap.String("orgId", orgId))
		planassignments, err := h.planSvc.ListOrgOrUserAssignments(c, orgId, "", paginationInput)
		if err != nil {
			h.logger.Error("failed to list assignments", zap.Reflect("error", err))
			errResp := domainerrors.NewErrorResponse(err)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
		return ctx.
			Status(fiber.StatusOK).JSON(models.NewHttpResponse(planassignments, "assignments retrieved successfully", fiber.StatusOK))
	}

	if userId != "" {
		// Call service to list org assignments
		planassignments, err := h.planSvc.ListOrgOrUserAssignments(c, "", userId, paginationInput)
		if err != nil {
			h.logger.Error("failed to list assignments", zap.Reflect("error", err))
			errResp := domainerrors.NewErrorResponse(err)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
		return ctx.
			Status(fiber.StatusOK).JSON(models.NewHttpResponse(planassignments, "assignments retrieved successfully", fiber.StatusOK))
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(planassignments, "assignments retrieved successfully", fiber.StatusOK))
}
