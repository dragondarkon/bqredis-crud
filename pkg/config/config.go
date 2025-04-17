package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	GoogleCloudProject string
	BigQueryDataset    string
	BigQueryTable      string
	RedisAddr          string
	RedisPassword      string
	RedisTTL           time.Duration
	Port               string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Set defaults and override with environment variables
	config := &Config{
		GoogleCloudProject: getEnv("GOOGLE_CLOUD_PROJECT", ""),
		BigQueryDataset:    getEnv("BIGQUERY_DATASET", "users_dataset"),
		BigQueryTable:      getEnv("BIGQUERY_TABLE", "users"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisTTL:           time.Duration(getEnvAsInt("REDIS_TTL_MINUTES", 5)) * time.Minute,
		Port:               getEnv("PORT", "8080"),
	}

	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return int(value.Minutes())
	}
	return defaultValue
}
