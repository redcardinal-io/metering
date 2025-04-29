package timeutil

import (
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"time"
)

func formatTimeUTC(t *time.Time, defaultValue string) string {
	if t == nil {
		return defaultValue
	}
	return t.UTC().Format(constants.TimeFormat)
}
