package assignments

import (
	"context"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"go.uber.org/zap"
)

// validateUUID checks if a string is a valid UUID
func validateUUID(id string) bool {
	if id == "" {
		return true // Empty is allowed for optional params
	}

	regex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return regex.MatchString(id)
}

// validateTimeFormat checks if a string is in the correct time format
func validateTimeFormat(t string) bool {
	if t == "" {
		return true // Empty is allowed for optional params
	}

	_, err := time.Parse(constants.TimeFormat, t)
	return err == nil
}

// @Summary List plan assignments
// @Description Get a list of plan assignments with optional filtering
// @Tags plan-assignments
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param valid_from query string false "Valid from date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param valid_until query string false "Valid until date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param plan_id_or_slug query string false "Plan ID or slug"
// @Param org_id query string false "Organization ID"
// @Param user_id query string false "User ID"
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
	validFrom := ctx.Query("valid_from")
	validUntil := ctx.Query("valid_until")
	planIdOrSlug := ctx.Query("plan_id_or_slug")

	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	// Validate UUID parameters
	orgID := ctx.Query("org_id")
	userID := ctx.Query("user_id")

	if orgID != "" && !validateUUID(orgID) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid organization ID format")
		h.logger.Error("invalid organization ID format", zap.String("org_id", orgID))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if userID != "" && !validateUUID(userID) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid user ID format")
		h.logger.Error("invalid user ID format", zap.String("user_id", userID))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Validate time format parameters
	if validFrom != "" && !validateTimeFormat(validFrom) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid valid_from timestamp format")
		h.logger.Error("invalid valid_from timestamp format", zap.String("valid_from", validFrom))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if validUntil != "" && !validateTimeFormat(validUntil) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid valid_until timestamp format")
		h.logger.Error("invalid valid_until timestamp format", zap.String("valid_until", validUntil))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if validFrom != "" && validUntil != "" && validFrom > validUntil {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "valid_from cannot be after valid_until")
		h.logger.Error("valid_from cannot be after valid_until", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

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
		OrganizationID: ctx.Query("org_id"),
		UserID:         ctx.Query("user_id"),
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
// @Param valid_from_before query string false "Valid from before date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param valid_from_after query string false "Valid from after date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param valid_until_before query string false "Valid until before date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param valid_until_after query string false "Valid until after date (format: YYYY-MM-DDThh:mm:ssZ)"
// @Param action query string false "Action type (Create, Update, Delete)"
// @Param plan_id_or_slug query string false "Plan ID or slug"
// @Param org_id query string false "Organization ID"
// @Param user_id query string false "User ID"
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
	validFromBefore := ctx.Query("valid_from_before")
	validFromAfter := ctx.Query("valid_from_after")
	validUntilBefore := ctx.Query("valid_until_before")
	validUntilAfter := ctx.Query("valid_until_after")
	action := ctx.Query("action")
	planIdOrSlug := ctx.Query("plan_id_or_slug")

	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	// Validate UUID parameters
	orgID := ctx.Query("org_id")
	userID := ctx.Query("user_id")

	if orgID != "" && !validateUUID(orgID) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid organization ID format")
		h.logger.Error("invalid organization ID format", zap.String("org_id", orgID))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if userID != "" && !validateUUID(userID) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid user ID format")
		h.logger.Error("invalid user ID format", zap.String("user_id", userID))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Validate time format parameters
	if validFromBefore != "" && !validateTimeFormat(validFromBefore) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid valid_from_before timestamp format")
		h.logger.Error("invalid valid_from_before timestamp format", zap.String("valid_from_before", validFromBefore))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if validFromAfter != "" && !validateTimeFormat(validFromAfter) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid valid_from_after timestamp format")
		h.logger.Error("invalid valid_from_after timestamp format", zap.String("valid_from_after", validFromAfter))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if validUntilBefore != "" && !validateTimeFormat(validUntilBefore) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid valid_until_before timestamp format")
		h.logger.Error("invalid valid_until_before timestamp format", zap.String("valid_until_before", validUntilBefore))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if validUntilAfter != "" && !validateTimeFormat(validUntilAfter) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid valid_until_after timestamp format")
		h.logger.Error("invalid valid_until_after timestamp format", zap.String("valid_until_after", validUntilAfter))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

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
		OrganizationID:   ctx.Query("org_id"),
		UserID:           ctx.Query("user_id"),
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
