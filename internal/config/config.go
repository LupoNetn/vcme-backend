package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	Port             string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

// LoadConfig initializes the configuration.
func LoadConfig() (*Config, error) {
	// Attempt to load .env file. We ignore the error because in production
	// environment variables might be set directly without a .env file.
	_ = godotenv.Load()

	dbURL, err := getRequiredEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	port := getEnvWithDefault("PORT", "8080")

	jwtAccessSecret, err := getRequiredEnv("JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}

	jwtRefreshSecret, err := getRequiredEnv("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL:      dbURL,
		Port:             port,
		JWTAccessSecret:  jwtAccessSecret,
		JWTRefreshSecret: jwtRefreshSecret,
	}, nil
}

// getRequiredEnv returns an error if the environment variable is not set.
func getRequiredEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("required environment variable %s not set", key)
	}
	return val, nil
}

// getEnvWithDefault returns a fallback value if the environment variable is not set.
func getEnvWithDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
