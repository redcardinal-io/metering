package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/interfaces/http/routes/constants"
	"go.uber.org/zap"
)

type event struct {
	ID           string         `json:"id" validate:"omitempty,uuid"`
	Type         string         `json:"type" validate:"required"`
	Source       string         `json:"source"`
	Organization string         `json:"organization" validate:"required"`
	User         string         `json:"user" validate:"required"`
	Timestamp    string         `json:"timestamp"`
	Properties   map[string]any `json:"properties"`
}

type publisEventRequestBody struct {
	Events              []event `json:"events" validate:"required,dive"`
	AllowPartialSuccess *bool   `json:"allow_partial_success" validate:"omitempty"`
}

func (h *httpHandler) publishEvent(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	if tenantSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EUNAUTHORIZED, fmt.Sprintf("header %s is required", constants.TenantHeader))
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var body publisEventRequestBody

	if err := ctx.BodyParser(&body); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Validate the request body
	if err := h.validator.Struct(body); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid request body")
		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if len(body.Events) == 0 {
		errResp := domainerrors.NewErrorResponseWithOpts(fmt.Errorf("cannot process empty event batch"), domainerrors.EINVALID, "empty event batch")
		h.logger.Error("empty event batch", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// default to true if not provided
	var allowPartialSuccess bool
	if body.AllowPartialSuccess == nil {
		allowPartialSuccess = true
	} else {
		allowPartialSuccess = *body.AllowPartialSuccess
	}

	events := models.EventBatch{}
	events.Events = make([]*models.Event, 0, len(body.Events))

	for _, event := range body.Events {
		if event.ID == "" {
			id, _ := uuid.NewV7()
			event.ID = id.String()
		}
		if event.Timestamp == "" {
			event.Timestamp = time.Now().Format(constants.TimeFormat)
		} else {
			_, err := time.Parse(constants.TimeFormat, event.Timestamp)
			if err != nil {
				errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid timestamp format")
				h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
				return ctx.Status(errResp.Status).JSON(errResp.ToJson())
			}
		}
		properties, err := json.Marshal(event.Properties)
		if err != nil {
			h.logger.Error("failed to parse properties", zap.Reflect("error", err))
			errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse properties")
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
		events.Events = append(events.Events, &models.Event{
			ID:           event.ID,
			Type:         event.Type,
			Source:       event.Source,
			Organization: event.Organization,
			User:         event.User,
			Timestamp:    event.Timestamp,
			Properties:   string(properties),
		})
	}

	res, err := h.producer.PublishEvents(context.Background(), h.publishTopic, &events, allowPartialSuccess)
	if err != nil {
		h.logger.Error("failed to publish events", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse[any](res, "events published successfully", fiber.StatusOK))
}
