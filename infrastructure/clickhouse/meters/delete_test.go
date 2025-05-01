package meters

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteMeter(t *testing.T) {
	tests := []struct {
		name        string
		deleteMeter DeleteMeter
		wantSQL     string
		wantArgs    []any
	}{
		{
			name: "Delete simple meter",
			deleteMeter: DeleteMeter{
				MeterSlug:  "page_views",
				TenantSlug: "test_tenant",
			},
			wantSQL:  "drop view if exists %s",
			wantArgs: []any{"rc_test_tenant_page_views_mv"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, _ := tt.deleteMeter.ToSQL()
			tt.wantSQL = fmt.Sprintf(tt.wantSQL, tt.wantArgs[0])
			// Compare SQL and args
			assert.Equal(t, tt.wantSQL, gotSQL)
		})
	}
}
