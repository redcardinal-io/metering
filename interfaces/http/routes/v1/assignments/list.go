package assignments

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"go.uber.org/zap"
)

func (h *httpHandler) list(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	var parsedValidFrom, parsedValidUntil time.Time
	var genErr error
	var planId *uuid.UUID
	var planassignments *pagination.PaginationView[models.PlanAssignment]
	validFrom := ctx.Query("validFrom")
	validUntil := ctx.Query("validUntil")
	planIdOrSlug := ctx.Query("planIdOrSlug")

	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	if len(ctx.Queries()) == 0 {
		// Call service to list all assignments
		planassignments, err := h.planSvc.ListAllAssignments(c, paginationInput)
		if err != nil {
			h.logger.Error("failed to list assignments", zap.Reflect("error", err))
			errResp := domainerrors.NewErrorResponse(err)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}

		return ctx.
			Status(fiber.StatusOK).JSON(models.NewHttpResponse(planassignments, "assignments retrieved successfully", fiber.StatusOK))
	}

	if validFrom != "" {
		parsedValidFrom, genErr = time.Parse(constants.TimeFormat, validFrom)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(genErr, domainerrors.EINVALID, "invalid timestamp format")
			h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	if validUntil != "" {
		parsedValidUntil, genErr = time.Parse(constants.TimeFormat, validUntil)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(genErr, domainerrors.EINVALID, "invalid timestamp format")
			h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	if planIdOrSlug != "" {
		planId, genErr = getPlanIDFromIdentifier(c, planIdOrSlug, h.planSvc)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(genErr, domainerrors.EINVALID, "invalid plan id or slug")
			h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	var queryAssignments = models.QueryPlanAssignmentInput{
		PlanID:         planId,
		OrganizationID: ctx.Query("orgId"),
		UserID:         ctx.Query("userId"),
		ValidFrom:      parsedValidFrom.UTC(),
		ValidUntil:     parsedValidUntil.UTC(),
	}

	if queryAssignments.OrganizationID != "" && queryAssignments.UserID != "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "provide only one profile param")
		h.logger.Error("provide only one profile param", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Call service to list assignments
	planassignments, err := h.planSvc.ListAssignments(c, queryAssignments, paginationInput)
	if err != nil {
		h.logger.Error("failed to list assignments", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(planassignments, "assignments retrieved successfully", fiber.StatusOK))
}
