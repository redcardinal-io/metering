package http

import (
	"fmt"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/interfaces/http/routes"
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

	// Register routes
	routes.RegisterRoutes(app, logger)

	// Start server
	return app.Listen(":" + config.Server.Port)
}
