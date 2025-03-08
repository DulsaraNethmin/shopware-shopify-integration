package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/lpernett/godotenv"
	"log"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
}

// ServerConfig holds server related configuration
type ServerConfig struct {
	Port   int
	Secret string
}

// DatabaseConfig holds database related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// AWSConfig holds AWS related configuration
type AWSConfig struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	StepFunctionsARN string
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found or could not be loaded. Using environment variables only.")
	}
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:   port,
			Secret: getEnv("SERVER_SECRET", "your-secret-key"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "shopware_shopify"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		AWS: AWSConfig{
			Region:           getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:      getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey:  getEnv("AWS_SECRET_ACCESS_KEY", ""),
			StepFunctionsARN: getEnv("AWS_STEP_FUNCTIONS_ARN", ""),
		},
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
