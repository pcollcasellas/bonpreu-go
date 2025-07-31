package config

import (
	"os"
	"strconv"
	"time"
)

// Configuration holds application configuration
type Configuration struct {
	SitemapURL      string
	RequestDuration time.Duration // Duration to spread requests over (0 = no rate limiting)
	HTTPClient      HTTPClientConfig
	Database        DatabaseConfig
}

// HTTPClientConfig holds HTTP client configuration
type HTTPClientConfig struct {
	Timeout int // in seconds
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntWithDefault gets an environment variable as int or returns a default value
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Configuration {
	return &Configuration{
		SitemapURL:      getEnvWithDefault("SITEMAP_URL", "https://www.compraonline.bonpreuesclat.cat/sitemaps/sitemap-products-part1.xml"),
		RequestDuration: time.Duration(getEnvIntWithDefault("REQUEST_DURATION_MINUTES", 1)) * time.Minute,
		HTTPClient: HTTPClientConfig{
			Timeout: getEnvIntWithDefault("HTTP_TIMEOUT_SECONDS", 30),
		},
		Database: DatabaseConfig{
			Host:     getEnvWithDefault("DB_HOST", "localhost"),
			Port:     getEnvIntWithDefault("DB_PORT", 5432),
			User:     getEnvWithDefault("DB_USER", ""),
			Password: getEnvWithDefault("DB_PASSWORD", ""),
			DBName:   getEnvWithDefault("DB_NAME", "bonpreu_db"),
			SSLMode:  getEnvWithDefault("DB_SSL_MODE", "require"),
		},
	}
}

// TestingConfig returns configuration for testing (no rate limiting)
func TestingConfig() *Configuration {
	return &Configuration{
		SitemapURL:      getEnvWithDefault("SITEMAP_URL", "https://www.compraonline.bonpreuesclat.cat/sitemaps/sitemap-products-part1.xml"),
		RequestDuration: 0, // No rate limiting for testing
		HTTPClient: HTTPClientConfig{
			Timeout: getEnvIntWithDefault("HTTP_TIMEOUT_SECONDS", 30),
		},
		Database: DatabaseConfig{
			Host:     getEnvWithDefault("DB_HOST", "localhost"),
			Port:     getEnvIntWithDefault("DB_PORT", 5432),
			User:     getEnvWithDefault("DB_USER", ""),
			Password: getEnvWithDefault("DB_PASSWORD", ""),
			DBName:   getEnvWithDefault("DB_NAME", "bonpreu_db"),
			SSLMode:  getEnvWithDefault("DB_SSL_MODE", "require"),
		},
	}
}
