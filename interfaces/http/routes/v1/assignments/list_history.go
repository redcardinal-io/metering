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

func (h *httpHandler) listhistory(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	var parsedValidFromBefore, parsedValidFromAfter, parsedValidUntilBefore, parsedValidUntilAfter time.Time
	var genErr error
	var planId *uuid.UUID
	var planassignments *pagination.PaginationView[models.PlanAssignmentHistory]
	validFromBefore := ctx.Query("validFromBefore")
	validFromAfter := ctx.Query("validFromAfter")
	validUntilBefore := ctx.Query("validUntilBefore")
	validUntilAfter := ctx.Query("validUntilAfter")
	action := ctx.Query("action")
	planIdOrSlug := ctx.Query("planIdOrSlug")

	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	if action == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "action filter is mandatory: Insert, Update and Delete")
		h.logger.Error("action filter is mandaotry: Insert, Update and Delete", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if !models.ValidateHistoryAction(action) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "provide valid action : Insert, Update and Delete")
		h.logger.Error("provide valid action : Insert, Update and Delete", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if validFromBefore != "" {
		parsedValidFromBefore, genErr = time.Parse(constants.TimeFormat, validFromBefore)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(genErr, domainerrors.EINVALID, "invalid timestamp format")
			h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	if validFromAfter != "" {
		parsedValidFromAfter, genErr = time.Parse(constants.TimeFormat, validFromAfter)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(genErr, domainerrors.EINVALID, "invalid timestamp format")
			h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	if validUntilBefore != "" {
		parsedValidUntilBefore, genErr = time.Parse(constants.TimeFormat, validUntilBefore)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(genErr, domainerrors.EINVALID, "invalid timestamp format")
			h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	if validUntilAfter != "" {
		parsedValidUntilAfter, genErr = time.Parse(constants.TimeFormat, validUntilAfter)
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

	queryAssignments := models.QueryPlanAssignmentHistoryInput{
		PlanID:           planId,
		OrganizationID:   ctx.Query("orgId"),
		UserID:           ctx.Query("userId"),
		Action:           models.HistoryActionEnum(action),
		ValidFromBefore:  parsedValidFromBefore,
		ValidFromAfter:   parsedValidFromAfter,
		ValidUntilBefore: parsedValidUntilBefore,
		ValidUntilAfter:  parsedValidUntilAfter,
	}

	if queryAssignments.OrganizationID != "" && queryAssignments.UserID != "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "provide only one profile param")
		h.logger.Error("provide only one profile param", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Call service to list assignments history
	planassignments, err := h.planSvc.ListAssignmentsHistory(c, queryAssignments, paginationInput)
	if err != nil {
		h.logger.Error("failed to list assignments", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(planassignments, "assignments history retrieved successfully", fiber.StatusOK))
}
