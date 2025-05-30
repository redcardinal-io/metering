package meters

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type queryMeterRequest struct {
	MeterSlug      string              `json:"meter_slug" validate:"required"`
	FilterGroupBy  map[string][]string `json:"filter_group_by"`
	From           *time.Time          `json:"from"`
	To             *time.Time          `json:"to"`
	GroupBy        []string            `json:"group_by"`
	WindowSize     *models.WindowSize  `json:"window_size"`
	WindowTimeZone *string             `json:"window_time_zone"`
}

// @Summary Query meter data
// @Description Query meter data with filters and grouping options
// @Tags meters
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param query body queryMeterRequest true "Query parameters"
// @Success 200 {object} models.HttpResponse[models.QueryMeterResult] "Meter queried successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Meter not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/meters/query [post]
func (h *httpHandler) query(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)

	var req queryMeterRequest
	if err := ctx.BodyParser(&req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if err := h.validator.Struct(req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid request body")
		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)
	result, err := h.meterSvc.QueryMeter(c, models.QueryMeterParams{
		MeterSlug:      req.MeterSlug,
		FilterGroupBy:  req.FilterGroupBy,
		From:           req.From,
		To:             req.To,
		GroupBy:        req.GroupBy,
		WindowSize:     req.WindowSize,
		WindowTimeZone: req.WindowTimeZone,
	})
	if err != nil {
		h.logger.Error("failed to query meter", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(result, "meter queried successfully", fiber.StatusOK))
}
