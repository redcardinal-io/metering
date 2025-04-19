package meters

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/interfaces/http/routes/constants"
	"go.uber.org/zap"
)

type queryMeterRequest struct {
	MeterSlug      string              `json:"meter_slug" validate:"required"`
	Organizations  []string            `json:"organizations"`
	Users          []string            `json:"users"`
	FilterGroupBy  map[string][]string `json:"filter_group_by"`
	From           *time.Time          `json:"from"`
	To             *time.Time          `json:"to"`
	GroupBy        []string            `json:"group_by"`
	WindowSize     *models.WindowSize  `json:"window_size"`
	WindowTimeZone *string             `json:"window_time_zone"`
}

func (h *httpHandler) query(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	if tenantSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EUNAUTHORIZED, fmt.Sprintf("header %s is required", constants.TenantHeader))
		h.logger.Error("failed to parse request body", zap.Any("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req queryMeterRequest
	if err := ctx.BodyParser(&req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
		h.logger.Error("failed to parse request body", zap.Any("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if err := h.validator.Struct(req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid request body")
		h.logger.Error("invalid request body", zap.Any("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	result, err := h.meterSvc.QueryMeter(ctx.UserContext(), models.QueryMeterInput{
		MeterSlug:      req.MeterSlug,
		Organizations:  req.Organizations,
		Users:          req.Users,
		FilterGroupBy:  req.FilterGroupBy,
		From:           req.From,
		To:             req.To,
		GroupBy:        req.GroupBy,
		WindowSize:     req.WindowSize,
		WindowTimeZone: req.WindowTimeZone,
		TenantSlug:     tenantSlug,
	})

	if err != nil {
		h.logger.Error("failed to query meter", zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(result, "meter queried successfully", fiber.StatusOK))
}
