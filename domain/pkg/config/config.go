package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	Logger LoggerConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type LogLevel string

const (
	INFO     LogLevel = "info"
	WARN     LogLevel = "warn"
	CRITICAL LogLevel = "critical"
	ERROR    LogLevel = "error"
)

type LoggerConfig struct {
	Level   LogLevel
	LogFile string
	Mode    string
}

func initializeViper() error {
	viper.SetEnvPrefix("rcmetering")
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// Allow viper to use environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		log.Println("Using environment variables")
		return nil
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_PORT", "8000")
	viper.SetDefault("LOGGER_LEVEL", string(INFO))
	viper.SetDefault("LOGGER_MODE", "dev")
}

func validateConfig() error {
	if viper.GetString("SERVER_HOST") == "" {
		return fmt.Errorf("server host is required")
	}
	if viper.GetString("SERVER_PORT") == "" {
		return fmt.Errorf("server port must be greater than 0")
	}
	if viper.GetString("LOGGER_LEVEL") == "" {
		return fmt.Errorf("logger level is required")
	}
	if viper.GetString("LOGGER_MODE") == "" {
		return fmt.Errorf("logger mode is required")
	}
	return nil
}

func LoadConfig() (*Config, error) {
	if err := initializeViper(); err != nil {
		return nil, err
	}

	setDefaults()

	if err := validateConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		Server: ServerConfig{
			Host: viper.GetString("SERVER_HOST"),
			Port: viper.GetString("SERVER_PORT"),
		},
		Logger: LoggerConfig{
			Level:   LogLevel(viper.GetString("LOGGER_LEVEL")),
			LogFile: viper.GetString("LOGGER_LOGFILE"),
			Mode:    viper.GetString("LOGGER_MODE"),
		},
	}

	return config, nil
}
