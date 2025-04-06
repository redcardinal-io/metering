package store

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"go.uber.org/zap"
)

type ClickHouseStore struct {
	driver driver.Conn
	logger *logger.Logger
}

func ClickHouseStoreRepository(logger *logger.Logger) repositories.OlapRepository {
	return &ClickHouseStore{
		logger: logger,
	}
}

// Connect - Connect to Postgres
func (store *ClickHouseStore) Connect(cfg *config.ClickHouseConfig) error {
	store.logger.Info("Connecting to ClickHouse", zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("database", cfg.Database))

	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
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

	if err != nil {
		store.logger.Error("Error connecting to ClickHouse", zap.Error(err))
		return err
	}

	if err := conn.Ping(ctx); err != nil {
		store.logger.Error("Error Pinging ClickHouse", zap.Error(err))
		return err
	}
	store.driver = conn
	store.logger.Info("Connected to CLickHouse")
	return nil
}

func (store *ClickHouseStore) Close() error {
	if store.driver != nil {
		store.driver.Close()
	}
	return nil
}

func (store *ClickHouseStore) GetDB() any {
	return store.driver
}
