package meters

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/domain/models"
)

type queryMeterRequest struct {
	MeterSlug      string              `json:"meter_slug"`
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
	tenantSlug := ctx.Get("tenant-slug")
	if tenantSlug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant slug is required",
		})
	}

	var req queryMeterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query meter",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": result,
	})

}
