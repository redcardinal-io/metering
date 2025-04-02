package logger

import (
	"fmt"

	"github.com/redcardinal-io/metering/domain/pkg/config"
	"go.uber.org/zap"
)

type LogLevel string

type Logger struct {
	*zap.Logger
}

func NewLogger(config *config.LoggerConfig) (*Logger, error) {

	var logger *zap.Logger
	var zapConfig zap.Config
	if config.Mode == "dev" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	outputs := []string{"stdout"}
	if config.LogFile != "" {
		outputs = append(outputs, config.LogFile)
	}
	zapConfig.OutputPaths = outputs
	zapConfig.DisableStacktrace = true

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{logger}, nil
}
