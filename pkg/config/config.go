package config

import (
	"os"
	"strconv"
	"time"
)

// Configuration holds all application configuration settings.
// It includes settings for sitemap URL, request rate limiting,
// HTTP client configuration, and database connection details.
type Configuration struct {
	SitemapURL      string
	RequestDuration time.Duration
	HTTPClient      HTTPClientConfig
	Database        DatabaseConfig
}

// HTTPClientConfig holds HTTP client configuration settings.
type HTTPClientConfig struct {
	Timeout int // Timeout in seconds
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// getEnvWithDefault retrieves an environment variable value or returns a default.
// It checks if the environment variable exists and is not empty,
// returning the environment value if present, otherwise the default value.
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntWithDefault retrieves an environment variable as an integer or returns a default.
// It attempts to parse the environment variable as an integer,
// returning the parsed value if successful, otherwise the default value.
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// DefaultConfig returns the default configuration for production use.
// This configuration includes rate limiting to be respectful to servers,
// with requests spread over the duration specified in REQUEST_DURATION_MINUTES.
// Database settings are loaded from environment variables with sensible defaults.
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

// TestingConfig returns a configuration optimized for testing and development.
// This configuration disables rate limiting (RequestDuration = 0) to allow
// for faster processing during testing. All other settings are identical to DefaultConfig.
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
