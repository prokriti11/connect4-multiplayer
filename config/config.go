// Package config provides application configuration.
package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// Kafka (optional)
	KafkaBrokers []string
	KafkaTopic   string

	// Game settings
	MatchTimeout    int // seconds
	ReconnectWindow int // seconds
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		KafkaTopic:      getEnv("KAFKA_TOPIC", "connect4-events"),
		MatchTimeout:    getEnvInt("MATCH_TIMEOUT", 10),
		ReconnectWindow: getEnvInt("RECONNECT_WINDOW", 30),
	}

	// Parse Kafka brokers
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		// Simple split - in production use proper parsing
		cfg.KafkaBrokers = []string{brokers}
	}

	return cfg
}

// getEnv returns env value or default
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt returns env value as int or default
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
