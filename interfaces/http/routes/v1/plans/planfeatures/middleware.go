package planfeatures

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

// TenantPlanFeatureMiddleware creates a middleware that checks if the plan and feature belong to the tenant
func (h *httpHandler) TenantPlanFeatureMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Get tenant slug from context
		tenantSlug := ctx.Get(constants.TenantHeader)
		// Get plan ID from URL parameter
		planIDStr := ctx.Params("planID")
		if planIDStr == "" {
			errResp := domainerrors.NewErrorResponseWithOpts(
				nil,
				domainerrors.EINVALID,
				"plan ID is required",
			)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}

		planID, err := uuid.Parse(planIDStr)
		if err != nil {
			errResp := domainerrors.NewErrorResponseWithOpts(
				err,
				domainerrors.EINVALID,
				"invalid plan ID format",
			)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}

		featureIDStr := ctx.Params("featureID")
		if featureIDStr != "" {
			featureID, err := uuid.Parse(featureIDStr)
			if err != nil {
				errResp := domainerrors.NewErrorResponseWithOpts(
					err,
					domainerrors.EINVALID,
					"invalid feature ID format",
				)
				return ctx.Status(errResp.Status).JSON(errResp.ToJson())
			}

			// Add tenant slug to context
			c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

			// Check if plan and feature belong to tenant
			belongs, err := h.planSvc.CheckPlanAndFeatureForTenant(c, planID, featureID)
			if err != nil {
				h.logger.Error("failed to check plan and feature for tenant",
					zap.String("tenant", tenantSlug),
					zap.String("plan_id", planID.String()),
					zap.String("feature_id", featureID.String()),
					zap.Error(err))
				errResp := domainerrors.NewErrorResponse(err)
				return ctx.Status(errResp.Status).JSON(errResp.ToJson())
			}

			if !belongs {
				errResp := domainerrors.NewErrorResponseWithOpts(
					nil,
					domainerrors.EUNAUTHORIZED,
					"plan feature not found or does not belong to tenant",
				)
				return ctx.Status(errResp.Status).JSON(errResp.ToJson())
			}
		}

		return ctx.Next()
	}
}
