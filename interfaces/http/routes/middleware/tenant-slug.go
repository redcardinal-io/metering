package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
)

// CheckTenantMiddleware creates a middleware that ensures the tenant header is present
func CheckTenantMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		tenantSlug := ctx.Get(constants.TenantHeader)
		if tenantSlug == "" {
			errResp := domainerrors.NewErrorResponseWithOpts(
				nil,
				domainerrors.EUNAUTHORIZED,
				fmt.Sprintf("header %s is required", constants.TenantHeader),
			)
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
		return ctx.Next()
	}
}
