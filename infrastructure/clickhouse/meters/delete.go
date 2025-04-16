package meters

import (
	"github.com/huandu/go-sqlbuilder"
)

type DeleteMeter struct {
	MeterSlug  string
	TenantSlug string
}

func (d *DeleteMeter) ToSQL() (string, []any) {
	viewName := GetMeterViewName(d.TenantSlug, d.MeterSlug)
	builder := sqlbuilder.Buildf("drop view if exists %s", viewName)
	return builder.Build()
}
