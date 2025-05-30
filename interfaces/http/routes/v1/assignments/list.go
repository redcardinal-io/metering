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

// @Summary List plan assignments
// @Description Get a list of plan assignments with optional filtering
// @Tags plan-assignments
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param validFrom query string false "Valid from date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param validUntil query string false "Valid until date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param planIdOrSlug query string false "Plan ID or slug"
// @Param orgId query string false "Organization ID"
// @Param userId query string false "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} models.HttpResponse[pagination.PaginationView[models.PlanAssignment]] "Assignments retrieved successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/assignments [get]
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

	queryAssignments := models.QueryPlanAssignmentInput{
		PlanID:         planId,
		OrganizationID: ctx.Query("orgId"),
		UserID:         ctx.Query("userId"),
		ValidFrom:      parsedValidFrom,
		ValidUntil:     parsedValidUntil,
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

// @Summary List plan assignment history
// @Description Get historical records of plan assignments with optional filtering
// @Tags plan-assignments
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param validFromBefore query string false "Valid from before date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param validFromAfter query string false "Valid from after date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param validUntilBefore query string false "Valid until before date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param validUntilAfter query string false "Valid until after date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param action query string false "Action type (Create, Update, Delete)"
// @Param planIdOrSlug query string false "Plan ID or slug"
// @Param orgId query string false "Organization ID"
// @Param userId query string false "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} models.HttpResponse[pagination.PaginationView[models.PlanAssignmentHistory]] "Assignment history retrieved successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/assignments/history [get]
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

	if action != "" && !models.ValidateHistoryAction(action) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "provide valid action : Create, Update and Delete")
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
		Action:           action,
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
