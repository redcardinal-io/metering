package http

import (
	"fmt"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/kafka"
	"github.com/redcardinal-io/metering/interfaces/http/routes"
	"github.com/redcardinal-io/metering/interfaces/http/routes/v1/events"
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
		AppName:       "gopie",
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

	// intialize services
	producerService := services.NewProducerService(producer, logger)

	// Register routes
	routes := routes.NewHTTPHandler(logger)
	routes.RegisterRoutes(app)

	// register v1 routes
	v1 := app.Group("/v1")
	eventsRoutes := events.NewHTTPHandler(events.HttpHandlerParams{
		PublishTopic: config.Kafka.KafkaRawEventsTopic,
		Producer:     producerService,
		Logger:       logger,
	})
	eventsRoutes.RegisterRoutes(v1)

	// Start server
	return app.Listen(":" + config.Server.Port)
}
