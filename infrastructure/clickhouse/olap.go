package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redcardinal-io/metering/application/repositories"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
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

// ClickHouseOlapRepository creates a new ClickHouse OLAP repository instance with the provided logger.
func ClickHouseOlapRepository(logger *logger.Logger) repositories.OlapRepository {
	return &ClickHouseOlap{
		logger: logger,
	}
}

func (olap *ClickHouseOlap) Connect(cfg *config.OlapConfig) error {
	olap.logger.Info("Connecting to ClickHouse", zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("database", cfg.Database))

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

	olap.db = sqlx.NewDb(conn, "clickhouse")

	if err := olap.db.Ping(); err != nil {
		return err
	}

	olap.logger.Info("Connected to ClickHouse")
	return nil
}

func (olap *ClickHouseOlap) CreateMeter(ctx context.Context, arg models.CreateMeterInput) error {
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
	olap.logger.Debug("Creating meter SQL", zap.String("sql", sql), zap.Any("args", args))

	if err != nil {
		return domainerrors.New(err,
			domainerrors.EOLAP,
			"Error generating meter creation SQL",
			domainerrors.
				WithOperation("ClickHouse.CreateMeter"),
		)
	}

	_, err = olap.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return MapError(err, "ClickHouse.CreateMeter")
	}

	olap.logger.Info("Created meter", zap.String("meter", meters.GetMeterViewName(arg.TenantSlug, arg.MeterSlug)))
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
		return nil, domainerrors.New(err,
			domainerrors.EOLAP,
			"Error generating meter query SQL",
			domainerrors.
				WithOperation("ClickHouse.QueryMeter"),
		)
	}
	olap.logger.Debug("Querying meter SQL", zap.String("sql", sql), zap.Any("args", args))

	rows, err := olap.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, MapError(err, "ClickHouse.QueryMeter")
	}
	defer rows.Close()

	var results []models.QueryMeterRow
	for rows.Next() {
		var row models.QueryMeterRow
		if err := rows.StructScan(&row); err != nil {
			return nil, MapError(err, "ClickHouse.QueryMeter")
		}
		results = append(results, row)
	}

	olap.logger.Info("Queried meter", zap.String("meter", meters.GetMeterViewName(input.TenantSlug, input.MeterSlug)))

	return &models.QueryMeterOutput{
		WindowStart: queryMeter.From,
		WindowEnd:   queryMeter.To,
		WindowSize:  queryMeter.WindowSize,
		Data:        results,
	}, nil
}

func (olap *ClickHouseOlap) Close() error {
	if olap.db != nil {
		return MapError(olap.db.Close(), "ClickHouse.Close")
	}
	return nil
}

func (olap *ClickHouseOlap) DeleteMeter(ctx context.Context, input models.DeleteMeterInput) error {
	deleteMeter := meters.DeleteMeter{
		TenantSlug: input.TenantSlug,
		MeterSlug:  input.MeterSlug,
	}

	sql, args := deleteMeter.ToSQL()

	_, err := olap.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return MapError(err, "ClickHouse.DeleteMeter")
	}

	olap.logger.Info("Deleted meter", zap.String("meter", meters.GetMeterViewName(input.TenantSlug, input.MeterSlug)))
	return nil
}

func (olap *ClickHouseOlap) GetDB() any {
	return olap.db
}
