package http

import (
	"fmt"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/redcardinal-io/metering/application/services"
	_ "github.com/redcardinal-io/metering/docs"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/clickhouse"
	"github.com/redcardinal-io/metering/infrastructure/kafka"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store/features"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store/meters"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store/planassignments"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store/planfeatures"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store/plans"
	"github.com/redcardinal-io/metering/infrastructure/postgres/store/quotas"
	"github.com/redcardinal-io/metering/interfaces/http/routes"
	"github.com/redcardinal-io/metering/interfaces/http/routes/middleware"
	"github.com/redcardinal-io/metering/interfaces/http/routes/v1/assignments"
	"github.com/redcardinal-io/metering/interfaces/http/routes/v1/events"
	featuresRoutes "github.com/redcardinal-io/metering/interfaces/http/routes/v1/features"
	meterRoutes "github.com/redcardinal-io/metering/interfaces/http/routes/v1/meters"
	planRoutes "github.com/redcardinal-io/metering/interfaces/http/routes/v1/plans"
	"go.uber.org/zap"
)

// @title RedCardinal Metering API
// @version 1.0.0
// @description API for metering service that handles event tracking, plans, features, and quotas
// @termsOfService http://swagger.io/terms/
// @host localhost:8080
// @BasePath /
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

	// Create new Fiber instance
	app := fiber.New(fiber.Config{
		AppName: "RedCardinal Metering API v1.0.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Handle the error
			logger.Error("error handling request", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})

	// Swagger documentation route
	app.Get("/swagger/*", swagger.New(swagger.Config{
		Title:        "RedCardinal Metering API",
		DocExpansion: "list",
	}))

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

	// initialize repositories
	meterStore := meters.NewPostgresMeterStoreRepository(store.GetDB(), logger)
	planStore := plans.NewPostgresPlanStoreRepository(store.GetDB(), logger)
	featureStore := features.NewPgFeatureStoreRepository(store.GetDB(), logger)
	planAssignmentsStore := planassignments.NewPostgresPlanAssignmentsStoreRepository(store.GetDB(), logger)
	planFeatureStore := planfeatures.NewPgPlanFeatureStoreRepository(store.GetDB(), logger)
	plannFeatureQuotaStore := quotas.NewPlanFeatureQuotaRepository(store.GetDB(), logger)

	// initialize services
	producerService := services.NewProducerService(producer, meterStore)
	meterService := services.NewMeterService(olap, meterStore)
	planMangementService := services.NewPlanService(
		planStore,
		featureStore,
		planFeatureStore,
		planAssignmentsStore,
		plannFeatureQuotaStore,
	)

	// Register routes
	routes := routes.NewHTTPHandler(logger)
	routes.RegisterRoutes(app)

	// register v1 routes
	v1 := app.Group("/v1").Use(middleware.CheckTenantMiddleware())
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

	// plan assignment routes
	assignmentsRoutes := assignments.NewHTTPHandler(logger, planMangementService)
	assignmentsRoutes.RegisterRoutes(v1)

	// plan routes
	planRoutes := planRoutes.NewHTTPHandler(logger, planMangementService)
	planRoutes.RegisterRoutes(v1)

	// feature routes
	featuresRoutes := featuresRoutes.NewHTTPHandler(logger, planMangementService)
	featuresRoutes.RegisterRoutes(v1)

	// Start server
	return app.Listen(":" + config.Server.Port)
}
