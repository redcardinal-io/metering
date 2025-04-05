package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type LogLevel string

const (
	INFO     LogLevel = "info"
	WARN     LogLevel = "warn"
	CRITICAL LogLevel = "critical"
	ERROR    LogLevel = "error"
)

type ServerConfig struct {
	Host string
	Port string
}

type LoggerConfig struct {
	Level   LogLevel
	LogFile string
	Mode    string
}

type KafkaConfig struct {
	KafkaBootstrapServers string
	KafkaRawEventsTopic   string
	KafkaSecurityProtocol string
	KafkaSaslMechanisms   string
	KafkaUsername         string
	KafkaPassword         string
}

type Config struct {
	Server ServerConfig
	Logger LoggerConfig
	Kafka  KafkaConfig
}

func initializeViper() error {
	// Load .env file
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		log.Println("Using environment variables only")
	} else {
		log.Println("Config file loaded successfully")
	}

	// No prefix, we'll use the full environment variable names
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	return nil
}

func setDefaults() {
	viper.SetDefault("RCMETERING_SERVER_HOST", "localhost")
	viper.SetDefault("RCMETERING_SERVER_PORT", "8000")
	viper.SetDefault("RCMETERING_LOGGER_LEVEL", string(INFO))
	viper.SetDefault("RCMETERING_LOGGER_MODE", "dev")
	viper.SetDefault("RCMETERING_LOGGER_LOGFILE", "rcmetering.log")
}

func validateConfig(config *Config) error {
	type validation struct {
		value string
		name  string
	}
	validations := []validation{
		{config.Server.Host, "server host"},
		{config.Server.Port, "server port"},
		{string(config.Logger.Level), "logger level"},
		{config.Logger.Mode, "logger mode"},
		{config.Kafka.KafkaBootstrapServers, "kafka bootstrap servers"},
		{config.Kafka.KafkaRawEventsTopic, "kafka raw events topic"},
	}

	for _, v := range validations {
		if v.value == "" {
			return fmt.Errorf("missing %s", v.name)
		}
	}
	return nil
}

func LoadConfig() (*Config, error) {
	if err := initializeViper(); err != nil {
		return nil, err
	}

	setDefaults()

	// Debug output - check full names
	log.Println("Checking 'RCMETERING_KAFKA_BOOTSTRAP_SERVERS' value:", viper.GetString("RCMETERING_KAFKA_BOOTSTRAP_SERVERS"))

	config := &Config{
		Server: ServerConfig{
			Host: viper.GetString("RCMETERING_SERVER_HOST"),
			Port: viper.GetString("RCMETERING_SERVER_PORT"),
		},
		Logger: LoggerConfig{
			Level:   LogLevel(viper.GetString("RCMETERING_LOGGER_LEVEL")),
			LogFile: viper.GetString("RCMETERING_LOGGER_LOGFILE"),
			Mode:    viper.GetString("RCMETERING_LOGGER_MODE"),
		},
		Kafka: KafkaConfig{
			KafkaBootstrapServers: viper.GetString("RCMETERING_KAFKA_BOOTSTRAP_SERVERS"),
			KafkaRawEventsTopic:   viper.GetString("RCMETERING_KAFKA_RAW_EVENTS_TOPIC"),
			KafkaSecurityProtocol: viper.GetString("RCMETERING_KAFKA_SECURITY_PROTOCOL"),
			KafkaSaslMechanisms:   viper.GetString("RCMETERING_KAFKA_SASL_MECHANISMS"),
			KafkaUsername:         viper.GetString("RCMETERING_KAFKA_USERNAME"),
			KafkaPassword:         viper.GetString("RCMETERING_KAFKA_PASSWORD"),
		},
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}
