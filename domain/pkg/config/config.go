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

type StoreConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

type OlapConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

type KafkaConfig struct {
	KafkaBootstrapServers string
	KafkaRawEventsTopic   string
	KafkaSecurityProtocol string
	KafkaSaslMechanisms   string
	KafkaUsername         string
	KafkaPassword         string
	KafkaQueueSize        int
	KafkaMaxRetries       int
	KafkaRetryBackoffMs   int
}

type Config struct {
	Server     ServerConfig
	Logger     LoggerConfig
	Kafka      KafkaConfig
	Postgres   StoreConfig
	ClickHouse OlapConfig
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
	viper.SetDefault("RCMETERING_KAFKA_QUEUE_SIZE", 1000)
	viper.SetDefault("RCMETERING_KAFKA_MAX_RETRIES", 3)
	viper.SetDefault("RCMETERING_KAFKA_RETRY_BACKOFF_MS", 100)
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
		{config.Postgres.Host, "postgres host"},
		{config.Postgres.Port, "postgres port"},
		{config.Postgres.Database, "postgres database"},
		{config.Postgres.User, "postgres user"},
		{config.Postgres.Password, "postgres password"},
		{config.ClickHouse.Host, "clickhouse host"},
		{config.ClickHouse.Port, "clickhouse port"},
		{config.ClickHouse.Database, "clickhouse database"},
		{config.ClickHouse.User, "clickhouse user"},
		{config.ClickHouse.Password, "clickhouse password"},
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
			KafkaQueueSize:        viper.GetInt("RCMETERING_KAFKA_QUEUE_SIZE"),
			KafkaMaxRetries:       viper.GetInt("RCMETERING_KAFKA_MAX_RETRIES"),
			KafkaRetryBackoffMs:   viper.GetInt("RCMETERING_KAFKA_RETRY_BACKOFF_MS"),
		},
		Postgres: StoreConfig{
			Host:     viper.GetString("RCMETERING_POSTGRES_HOST"),
			Port:     viper.GetString("RCMETERING_POSTGRES_PORT"),
			Database: viper.GetString("RCMETERING_POSTGRES_DATABASE"),
			User:     viper.GetString("RCMETERING_POSTGRES_USER"),
			Password: viper.GetString("RCMETERING_POSTGRES_PASSWORD"),
		},
		ClickHouse: OlapConfig{
			Host:     viper.GetString("RCMETERING_CLICKHOUSE_HOST"),
			Port:     viper.GetString("RCMETERING_CLICKHOUSE_PORT"),
			Database: viper.GetString("RCMETERING_CLICKHOUSE_DATABASE"),
			User:     viper.GetString("RCMETERING_CLICKHOUSE_USER"),
			Password: viper.GetString("RCMETERING_CLICKHOUSE_PASSWORD"),
		},
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}
