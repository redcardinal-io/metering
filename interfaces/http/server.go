package http

import (
	"fmt"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/clickhouse"
	"github.com/redcardinal-io/metering/infrastructure/kafka"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store/meters"
	"github.com/redcardinal-io/metering/interfaces/http/routes"
	"github.com/redcardinal-io/metering/interfaces/http/routes/v1/events"
	meterRoutes "github.com/redcardinal-io/metering/interfaces/http/routes/v1/meters"
	"go.uber.org/zap"
)

func ServeHttp() error {
	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Initialize logger
	logger, err := logger.NewLogger(&config.Logger)
	if err != nil {
		return fmt.Errorf("error creating logger: %w", err)
	}
	logger.Info("logger initialized")
	logger.Info("rcmetering server starting at ",
		zap.String("host", config.Server.Host),
		zap.String("port", config.Server.Port))

	// Set up Fiber app
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		AppName:       "rcmetering",
	})

	// Configure middleware
	app.Use(cors.New())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger.Sugar().Desugar(),
	}))

	// initialize repositories
	producer, err := kafka.NewKafkaProducerRepository(logger, config.Kafka)
	if err != nil {
		return fmt.Errorf("error creating Kafka producer: %w", err)
	}
	defer producer.Close()

	// initialize OLAP repository
	olap := clickhouse.ClickHouseOlapRepository(logger)
	err = olap.Connect(&config.ClickHouse)
	if err != nil {
		return fmt.Errorf("error connecting to ClickHouse: %w", err)
	}
	defer olap.Close()

	// initialize store repository
	store := store.NewPostgresStoreRepository(logger)
	err = store.Connect(&config.Postgres)
	if err != nil {
		return fmt.Errorf("error connecting to Postgres: %w", err)
	}
	defer store.Close()
	meterStore := meters.NewPostgresMeterStoreRepository(store.GetDB(), logger)

	// intialize services
	producerService := services.NewProducerService(producer, logger)
	meterService := services.NewMeterService(olap, meterStore)

	// Register routes
	routes := routes.NewHTTPHandler(logger)
	routes.RegisterRoutes(app)

	// register v1 routes
	v1 := app.Group("/v1")
	// events routes
	eventsRoutes := events.NewHTTPHandler(events.HttpHandlerParams{
		PublishTopic: config.Kafka.KafkaRawEventsTopic,
		Producer:     producerService,
		Logger:       logger,
	})
	eventsRoutes.RegisterRoutes(v1)

	// meter routes
	meterRoutes := meterRoutes.NewHTTPHandler(logger, meterService)
	meterRoutes.RegisterRoutes(v1)

	// Start server
	return app.Listen(":" + config.Server.Port)
}
