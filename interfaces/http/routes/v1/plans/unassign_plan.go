package plans

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type UnAssignPlanRequest struct {
	OrganizationId string `json:"organization_id"`
	UserId         string `json:"user_id"`
}

func (h *httpHandler) unassignPlan(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	var req UnAssignPlanRequest

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
