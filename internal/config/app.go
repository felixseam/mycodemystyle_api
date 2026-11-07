package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	AppName     string
	AppEnv      string
	AppHost     string
	AppPort     int
	AppTimezone *time.Location
}

func NewAppConfig() *AppConfig {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using environment variables")
	}

	timezoneStr := getEnv("APP_TIMEZONE", "Asia/Phnom_Penh")
	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		log.Fatalf("Invalid APP_TIMEZONE %q: %v", timezoneStr, err)
	}

	return &AppConfig{
		AppName:     getEnv("APP_NAME", "cinema-backend"),
		AppEnv:      getEnv("APP_ENV", "development"),
		AppHost:     getEnv("APP_HOST", "0.0.0.0"),
		AppPort:     getEnvInt("APP_PORT", 8085),
		AppTimezone: loc,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Invalid integer value for %s: %v", key, err)
	}

	return port
}

func LoadTimezone(timezoneStr string) (*time.Location, error) {
	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		return nil, fmt.Errorf("load application timezone %q: %w", timezoneStr, err)
	}

	return loc, nil
}
