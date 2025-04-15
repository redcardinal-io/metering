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

type ClickHouseStore struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func ClickHouseOlapRepository(logger *logger.Logger) repositories.OlapRepository {
	return &ClickHouseStore{
		logger: logger,
	}
}

func (store *ClickHouseStore) CreateMeter(ctx context.Context, arg models.CreateMeterInput) error {
	createMeter := meters.CreateMeter{
		Slug:          arg.Slug,
		TenantSlug:    arg.TenantSlug,
		ValueProperty: arg.ValueProperty,
		Populate:      arg.Populate,
		Properties:    arg.Properties,
		Aggregation:   arg.Aggregation,
		EventType:     arg.EventType,
	}

	sql, args, err := createMeter.ToCreateSQL()
	fmt.Printf("SQL: %v\n", sql)
	for i, arg := range args {
		fmt.Printf("Arg-%v: %v\n", i, arg)
	}
	if err != nil {
		store.logger.Error("Error creating meter SQL", zap.Error(err))
		return err
	}

	_, err = store.db.ExecContext(ctx, sql, args...)
	if err != nil {
		store.logger.Error("Error executing meter creation SQL", zap.Error(err))
		return err
	}

	store.logger.Debug("Created meter", zap.String("slug", arg.Slug), zap.String("tenant_slug", arg.TenantSlug))
	return nil
}

func (store *ClickHouseStore) Connect(cfg *config.OlapConfig) error {
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

func (store *ClickHouseStore) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

func (store *ClickHouseStore) GetDB() any {
	return store.db
}
