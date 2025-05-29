package models

type MeteredResetPeriod string

const (
	MeteredResetPeriodDay    MeteredResetPeriod = "day"
	MeteredResetPeriodWeek   MeteredResetPeriod = "week"
	MeteredResetPeriodMonth  MeteredResetPeriod = "month"
	MeteredResetPeriodYear   MeteredResetPeriod = "year"
	MeteredResetPeriodCustom MeteredResetPeriod = "custom"
	MeteredResetPeriodRolling MeteredResetPeriod = "rolling"
	MeteredResetPeriodNever  MeteredResetPeriod = "never"
)

type MeteredActionAtLimit string

const (
	MeteredActionAtLimitNone     MeteredActionAtLimit = "none"
	MeteredActionAtLimitBlock    MeteredActionAtLimit = "block"
	MeteredActionAtLimitThrottle MeteredActionAtLimit = "throttle"
)

type PlanFeatureQuota struct {
	Base
	PlanFeatureID       string              `json:"plan_feature_id"`
	LimitValue         int64               `json:"limit_value"`
	ResetPeriod        MeteredResetPeriod  `json:"reset_period"`
	CustomPeriodMinutes *int64             `json:"custom_period_minutes,omitempty"`
	ActionAtLimit      MeteredActionAtLimit `json:"action_at_limit"`
}

type CreatePlanFeatureQuotaInput struct {
	PlanFeatureID       string              `json:"plan_feature_id" validate:"required,uuid"`
	LimitValue         int64               `json:"limit_value" validate:"required,gt=0"`
	ResetPeriod        MeteredResetPeriod  `json:"reset_period" validate:"required,oneof=day week month year custom rolling never"`
	CustomPeriodMinutes *int64             `json:"custom_period_minutes,omitempty" validate:"required_if=ResetPeriod custom,omitempty,gt=0"`
	ActionAtLimit      MeteredActionAtLimit `json:"action_at_limit" validate:"required,oneof=none block throttle"`
	CreatedBy          string              `json:"created_by" validate:"required"`
}

type UpdatePlanFeatureQuotaInput struct {
	PlanFeatureID       string               `json:"-" validate:"required,uuid"`
	LimitValue         *int64               `json:"limit_value,omitempty" validate:"omitempty,gt=0"`
	ResetPeriod        *MeteredResetPeriod  `json:"reset_period,omitempty" validate:"omitempty,oneof=day week month year custom rolling never"`
	CustomPeriodMinutes *int64              `json:"custom_period_minutes,omitempty" validate:"omitempty,gt=0"`
	ActionAtLimit      *MeteredActionAtLimit `json:"action_at_limit,omitempty" validate:"omitempty,oneof=none block throttle"`
	UpdatedBy          string               `json:"updated_by" validate:"required"`
}
