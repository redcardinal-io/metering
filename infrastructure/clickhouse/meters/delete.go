package meters

import (
	"github.com/huandu/go-sqlbuilder"
)

type deleteMeter struct {
	MeterSlug  string
	TenantSlug string
}

func (d *deleteMeter) DeleteMeter() (string, []any) {
	viewName := getMeterViewName(d.TenantSlug, d.MeterSlug)
	builder := sqlbuilder.Buildf("drop view if exists %s", viewName)
	return builder.Build()
}
