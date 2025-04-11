package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/redcardinal-io/metering/domain/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel defines the severity of log messages
type LogLevel string

const (
	// DebugLevel for development debugging
	DebugLevel LogLevel = "debug"
	// InfoLevel for general information
	InfoLevel LogLevel = "info"
	// WarnLevel for warnings
	WarnLevel LogLevel = "warn"
	// ErrorLevel for errors
	ErrorLevel LogLevel = "error"
	// PanicLevel for panic-worthy problems
	PanicLevel LogLevel = "panic"
	// FatalLevel for fatal errors
	FatalLevel LogLevel = "fatal"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
}

// ConvertLogLevel converts a LogLevel to a zapcore.Level
func ConvertLogLevel(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// NewLogger creates a new Logger instance
func NewLogger(config *config.LoggerConfig) (*Logger, error) {
	// Set log level
	level := zapcore.InfoLevel
	if config.Level != "" {
		level = ConvertLogLevel(LogLevel(config.Level))
	}

	// Configure encoders
	var consoleEncoder zapcore.Encoder

	if config.Mode == "dev" {
		// Development: colored console output
		encConfig := zap.NewDevelopmentEncoderConfig()
		encConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
		encConfig.EncodeCaller = zapcore.ShortCallerEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(encConfig)
	} else {
		// Production: JSON output
		encConfig := zap.NewProductionEncoderConfig()
		encConfig.TimeKey = "timestamp"
		encConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
		encConfig.EncodeCaller = zapcore.ShortCallerEncoder
		consoleEncoder = zapcore.NewJSONEncoder(encConfig)
	}

	// Setup cores
	var cores []zapcore.Core

	// Always add stdout
	stdoutSyncer := zapcore.AddSync(os.Stdout)
	cores = append(cores, zapcore.NewCore(consoleEncoder, stdoutSyncer, level))

	// Add file logging if configured
	if config.LogFile != "" {
		// Configure file encoder (always JSON for files)
		fileEncConfig := zap.NewProductionEncoderConfig()
		fileEncConfig.TimeKey = "timestamp"
		fileEncConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
		fileEncConfig.EncodeCaller = zapcore.ShortCallerEncoder
		fileEncoder := zapcore.NewJSONEncoder(fileEncConfig)

		// Set up log rotation
		fileWriter := &lumberjack.Logger{
			Filename:   config.LogFile,
			MaxSize:    100, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		}
		fileSyncer := zapcore.AddSync(fileWriter)
		cores = append(cores, zapcore.NewCore(fileEncoder, fileSyncer, level))
	}

	// Create the core and logger
	core := zapcore.NewTee(cores...)

	// Build options
	zapOptions := []zap.Option{
		zap.AddCaller(),
	}

	// In production mode, add stacktraces for errors
	if config.Mode != "dev" {
		zapOptions = append(zapOptions, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	zapLogger := zap.New(core, zapOptions...)

	return &Logger{zapLogger}, nil
}

// WithField returns a new logger with the field added to the logging context
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{l.Logger.With(zap.Any(key, value))}
}

// WithFields returns a new logger with multiple fields added to the logging context
func (l *Logger) WithFields(fields map[string]any) *Logger {
	if len(fields) == 0 {
		return l
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return &Logger{l.Logger.With(zapFields...)}
}

// WithError returns a new logger with the error added to the logging context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.Logger.With(zap.Error(err))}
}

// Debugf logs a formatted message at debug level
func (l *Logger) Debugf(format string, args ...any) {
	l.Logger.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted message at info level
func (l *Logger) Infof(format string, args ...any) {
	l.Logger.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted message at warn level
func (l *Logger) Warnf(format string, args ...any) {
	l.Logger.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted message at error level
func (l *Logger) Errorf(format string, args ...any) {
	l.Logger.Error(fmt.Sprintf(format, args...))
}

// Fatalf logs a formatted message at fatal level and then calls os.Exit(1)
func (l *Logger) Fatalf(format string, args ...any) {
	l.Logger.Fatal(fmt.Sprintf(format, args...))
}

// Panicf logs a formatted message at panic level and then panics
func (l *Logger) Panicf(format string, args ...any) {
	l.Logger.Panic(fmt.Sprintf(format, args...))
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
