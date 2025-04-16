package meters

import (
	"fmt"
	"strings"

	"github.com/redcardinal-io/metering/domain/models"
)

const eventsTable = "rc_events"

var aggregationMap = map[models.AggregationEnum]struct {
	stateFunc string
	mergeFunc string
	dataType  string
}{
	models.AggregationSum:         {"sumState", "sum", "Float64"},
	models.AggregationAvg:         {"avgState", "avg", "Float64"},
	models.AggregationMin:         {"minState", "min", "Float64"},
	models.AggregationMax:         {"maxState", "max", "Float64"},
	models.AggregationCount:       {"countState", "count", "Float64"},
	models.AggregationUniqueCount: {"uniqState", "uniq", "String"},
}

func GetMeterViewName(organization, meterSlug string) string {
	return fmt.Sprintf("rc_%s_%s_mv", organization, meterSlug)
}

func normalizeSQL(sql string) string {
	// Remove extra whitespace and standardize format
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")
	for strings.Contains(sql, "  ") {
		sql = strings.ReplaceAll(sql, "  ", " ")
	}
	return strings.TrimSpace(sql)
}
