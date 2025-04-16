package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/clickhouse/meters"
	"go.uber.org/zap"
)

type ClickHouseOlap struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func ClickHouseOlapRepository(logger *logger.Logger) repositories.OlapRepository {
	return &ClickHouseOlap{
		logger: logger,
	}
}

func (store *ClickHouseOlap) Connect(cfg *config.OlapConfig) error {
	store.logger.Info("Connecting to ClickHouse", zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("database", cfg.Database))

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "RedCardinal", Version: "0.1"},
			},
		},
	})

	store.db = sqlx.NewDb(conn, "clickhouse")

	if err := store.db.Ping(); err != nil {
		store.logger.Error("Error pinging ClickHouse", zap.Error(err))
		return err
	}

	store.logger.Info("Connected to ClickHouse")
	return nil
}

func (store *ClickHouseOlap) CreateMeter(ctx context.Context, arg models.CreateMeterInput) error {
	createMeter := meters.CreateMeter{
		Slug:          arg.MeterSlug,
		TenantSlug:    arg.TenantSlug,
		ValueProperty: arg.ValueProperty,
		Populate:      arg.Populate,
		Properties:    arg.Properties,
		Aggregation:   arg.Aggregation,
		EventType:     arg.EventType,
	}

	sql, args, err := createMeter.ToCreateSQL()
	store.logger.Debug("Creating meter SQL", zap.String("sql", sql), zap.Any("args", args))

	if err != nil {
		store.logger.Error("Error creating meter SQL", zap.Error(err))
		return err
	}

	_, err = store.db.ExecContext(ctx, sql, args...)
	if err != nil {
		store.logger.Error("Error executing meter creation SQL", zap.Error(err))
		return err
	}

	store.logger.Info("Created meter", zap.String("meter", meters.GetMeterViewName(arg.TenantSlug, arg.MeterSlug)))
	return nil
}

func (olap *ClickHouseOlap) QueryMeter(ctx context.Context, input models.QueryMeterInput, agg *models.AggregationEnum) (*models.QueryMeterOutput, error) {
	queryMeter := meters.QueryMeter{
		TenantSlug:     input.TenantSlug,
		MeterSlug:      input.MeterSlug,
		Organizations:  input.Organizations,
		Users:          input.Users,
		FilterGroupBy:  input.FilterGroupBy,
		From:           input.From,
		To:             input.To,
		GroupBy:        input.GroupBy,
		WindowSize:     input.WindowSize,
		WindowTimeZone: input.WindowTimeZone,
		Aggregation:    *agg,
	}

	sql, args, err := queryMeter.ToSQL()
	if err != nil {
		olap.logger.Error("Error generating query SQL", zap.Error(err))
		return nil, err
	}
	olap.logger.Debug("Querying meter SQL", zap.String("sql", sql), zap.Any("args", args))

	rows, err := olap.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		olap.logger.Error("Error executing query SQL", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var results []models.QueryMeterRow
	for rows.Next() {
		var row models.QueryMeterRow
		if err := rows.StructScan(&row); err != nil {
			olap.logger.Error("Error scanning row", zap.Error(err))
			return nil, err
		}
		results = append(results, row)
	}

	return &models.QueryMeterOutput{
		WindowStart: queryMeter.From,
		WindowEnd:   queryMeter.To,
		WindowSize:  queryMeter.WindowSize,
		Data:        results,
	}, nil
}

func (store *ClickHouseOlap) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

func (store *ClickHouseOlap) DeleteMeter(ctx context.Context, input models.DeleteMeterInput) error {
	deleteMeter := meters.DeleteMeter{
		TenantSlug: input.TenantSlug,
		MeterSlug:  input.MeterSlug,
	}

	sql, args := deleteMeter.ToSQL()

	_, err := store.db.ExecContext(ctx, sql, args...)
	if err != nil {
		store.logger.Error("Error executing delete meter SQL", zap.Error(err))
		return err
	}

	store.logger.Info("Deleted meter", zap.String("meter", meters.GetMeterViewName(input.TenantSlug, input.MeterSlug)))
	return nil
}

func (store *ClickHouseOlap) GetDB() any {
	return store.db
}
