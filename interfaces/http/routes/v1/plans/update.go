package plans

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type updatePlanRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description string `json:"description,omitempty" validate:"omitempty,min=10,max=255"`
	UpdatedBy   string `json:"updated_by" validate:"required,min=3,max=255"`
}

func (h *httpHandler) update(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan ID is required")
		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req updatePlanRequest
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

	plan, err := h.planSvc.UpdatePlanByIDorSlug(c, idOrSlug, models.UpdatePlanInput{
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   req.UpdatedBy,
	})
	if err != nil {
		h.logger.Error("failed to update plan", zap.String("idOrSlug", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(plan, "plan updated successfully", fiber.StatusOK))
}

type terminatePlanRequest struct {
	OrganizationId string `json:"organization_id"`
	UserId         string `json:"user_id"`
}

func (h *httpHandler) terminatePlan(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	var req terminatePlanRequest

	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan ID is required")
		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

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

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenant_slug)

	planId, getErr := getPlanId(h, c, idOrSlug)

	if getErr != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(getErr, domainerrors.EINVALID, "invalid plan id or slug")
		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.OrganizationId != "" && req.UserId != "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "organization_id and user_id are mutually exclusive, provide any one")
		h.logger.Error("organization_id and user_id are mutually exclusive, provide any one", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var isOrg bool
	var orgOrUserId uuid.UUID
	var genErr error

	if req.OrganizationId != "" {
		isOrg = true
		orgOrUserId, genErr = uuid.Parse(req.OrganizationId)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "unable to parse organization_id")
			h.logger.Error("unable to parse organization_id", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	} else {
		isOrg = false
		orgOrUserId, genErr = uuid.Parse(req.UserId)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "unable to parse user_id")
			h.logger.Error("unable to parse user_id", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	genErr = h.planSvc.UnAssignPlan(c, *planId, orgOrUserId, isOrg)
	if genErr != nil {
		h.logger.Error("failed to un-assign plan", zap.Reflect("error", genErr))
		errResp := domainerrors.NewErrorResponse(genErr)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		SendStatus(fiber.StatusNoContent)
}

type upadateAssignedPlanRequest struct {
	OrganizationId string  `json:"organization_id"`
	UserId         string  `json:"user_id"`
	ValidFrom      *string `json:"valid_from" validate:"required"`
	ValidUntil     *string `json:"valid_until"`
	UpdatedBy      string  `json:"updated_by" validate:"required"`
}

func (h *httpHandler) updateAssignedPlan(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	var req upadateAssignedPlanRequest

	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan ID is required")
		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

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

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenant_slug)

	planId, getErr := getPlanId(h, c, idOrSlug)

	if getErr != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(getErr, domainerrors.EINVALID, "invalid plan id or slug")
		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.OrganizationId != "" && req.UserId != "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "organization_id and user_id are mutually exclusive, provide any one")
		h.logger.Error("organization_id and user_id are mutually exclusive, provide any one", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var timeFormaterr error
	var isOrg bool
	var orgOrUserId uuid.UUID
	var genErr error
	var planAssignment *models.PlanAssignment

	if req.OrganizationId != "" {
		isOrg = true
		orgOrUserId, genErr = uuid.Parse(req.OrganizationId)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "unable to parse organization_id")
			h.logger.Error("unable to parse organization_id", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	} else {
		isOrg = false
		orgOrUserId, genErr = uuid.Parse(req.UserId)
		if genErr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "unable to parse user_id")
			h.logger.Error("unable to parse user_id", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	ValidFrom, timeFormaterr := time.Parse(constants.TimeFormat, *req.ValidFrom)
	if timeFormaterr != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(timeFormaterr, domainerrors.EINVALID, "invalid timestamp format")
		h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}
	if req.ValidUntil == nil {
		var ValidUntil pgtype.Timestamptz

		planAssignment, genErr = h.planSvc.UpdateAssignedPlan(c, *planId, models.AssignOrUpdateAssignedPlanInput{
			OrganizationOrUserId: pgtype.UUID{Bytes: orgOrUserId, Valid: true},
			ValidFrom:            pgtype.Timestamptz{Time: ValidFrom, Valid: true},
			ValidUntil:           ValidUntil,
			By:                   req.UpdatedBy,
		}, isOrg)

		if genErr != nil {
			h.logger.Error("failed to update assigned plan", zap.Reflect("error", genErr))
			errResp := domainerrors.NewErrorResponse(genErr)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	} else {
		ValidUntil, timeFormaterr := time.Parse(constants.TimeFormat, *req.ValidUntil)
		if timeFormaterr != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(timeFormaterr, domainerrors.EINVALID, "invalid timestamp format")
			h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}

		planAssignment, genErr = h.planSvc.UpdateAssignedPlan(c, *planId, models.AssignOrUpdateAssignedPlanInput{
			OrganizationOrUserId: pgtype.UUID{Bytes: orgOrUserId, Valid: true},
			ValidFrom:            pgtype.Timestamptz{Time: ValidFrom, Valid: true},
			ValidUntil:           pgtype.Timestamptz{Time: ValidUntil, Valid: true},
			By:                   req.UpdatedBy,
		}, isOrg)

		if genErr != nil {
			h.logger.Error("failed to update assigned plan", zap.Reflect("error", genErr))
			errResp := domainerrors.NewErrorResponse(genErr)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	return ctx.
		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(planAssignment, "updated assignment successfully", fiber.StatusCreated))
}

type archivePlanRequest struct {
	UpdatedBy string `json:"updated_by" validate:"required,min=3,max=255"`
	Archive   bool   `json:"archive"`
}

func (h *httpHandler) archive(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "Plan Id or Slug is required")
		h.logger.Error("Plan Id or Slug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req archivePlanRequest
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

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	err := h.planSvc.ArchivePlanByIDorSlug(c, idOrSlug, models.ArchivePlanInput{
		UpdatedBy: req.UpdatedBy,
		Archive:   req.Archive,
	})
	if err != nil {
		h.logger.Error("failed to toggle archive plan", zap.String("id", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		SendStatus(fiber.StatusNoContent)
}
