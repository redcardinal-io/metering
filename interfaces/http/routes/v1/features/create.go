package features

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type createFeatureRequest struct {
	Name        string         `json:"name" validate:"required"`
	Description string         `json:"description" validate:"required"`
	Slug        string         `json:"slug" validate:"required"`
	Type        string         `json:"type" validate:"required,oneof=standard metered"`
	Config      map[string]any `json:"config" validate:"required"`
	CreatedBy   string         `json:"created_by" validate:"required"`
}

func (h *httpHandler) create(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)

	var req createFeatureRequest
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

	feature, err := h.featureSvc.CreateFeature(c, models.CreateFeatureInput{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
		Type:        models.FeatureTypeEnum(req.Type),
		TenantSlug:  tenant_slug,
		Config:      req.Config,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		h.logger.Error("failed to create feature", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(feature, "feature created successfully", fiber.StatusCreated))
}
