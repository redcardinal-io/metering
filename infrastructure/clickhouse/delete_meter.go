package clickhouse

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/redcardinal-io/metering/domain/models"
	"go.uber.org/zap"
)

// DeleteMeter deletes a materialized view for meter data in ClickHouse
func (store *ClickHouseStore) DeleteMeter(ctx context.Context, organization string, meterSlug string) error {
	viewName := getMeterViewName(organization, meterSlug)
	sql := fmt.Sprintf("DROP VIEW %s", viewName)
	
	_, err := store.db.ExecContext(ctx, sql)
	if err != nil {
		store.logger.Error("Failed to delete meter view", 
			zap.Error(err), 
			zap.String("namespace", organization),
			zap.String("meter_slug", meterSlug))
		return fmt.Errorf("failed to delete meter view: %w", err)
	}
	
	store.logger.Info("Successfully deleted meter view",
		zap.String("", viewName))
	return nil
}