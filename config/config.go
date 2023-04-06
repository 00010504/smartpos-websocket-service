package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	Environment string
	ServiceName string

	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string

	// services
	MarketingService    string
	InventoryService    string
	NotificationService string
	OrderService        string
	ReportService       string
	CatalogService      string

	LogLevel string
	HTTPHost string
	HTTPPort int

	KafkaUrl string
}

func Load() Config {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("No .env file found")
	}

	config := Config{}

	config.Environment = cast.ToString(getOrReturnDefault("ENVIRONMENT", "develop"))
	config.LogLevel = cast.ToString(getOrReturnDefault("LOG_LEVEL", "info"))
	config.ServiceName = cast.ToString(getOrReturnDefault("SERVICE_NAME", ""))
	config.HTTPHost = cast.ToString(getOrReturnDefault("HTTP_HOST", "0.0.0.0"))
	config.HTTPPort = cast.ToInt(getOrReturnDefault("HTTP_PORT", "8000"))

	config.PostgresHost = cast.ToString(getOrReturnDefault("POSTGRES_HOST", "localhost"))
	config.PostgresPort = cast.ToInt(getOrReturnDefault("POSTGRES_PORT", 5432))
	config.PostgresUser = cast.ToString(getOrReturnDefault("POSTGRES_USER", "postgres"))
	config.PostgresPassword = cast.ToString(getOrReturnDefault("POSTGRES_PASSWORD", "root"))
	config.PostgresDatabase = cast.ToString(getOrReturnDefault("POSTGRES_DATABASE", ""))

	config.KafkaUrl = cast.ToString(getOrReturnDefault("KAFKA_URL", "localhost:9092"))

	return config
}

func getOrReturnDefault(key string, defaultValue interface{}) interface{} {
	_, exists := os.LookupEnv(key)

	if exists {
		return os.Getenv(key)
	}

	return defaultValue
}
