package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	CoinGecko CoinGeckoConfig
	Worker   WorkerConfig
}

type ServerConfig struct {
	HTTPPort string
	GRPCPort string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	TTL      time.Duration // Cache TTL in seconds
}

type CoinGeckoConfig struct {
	APIKey     string
	BaseURL    string
	RateLimit  int // Requests per minute
	Timeout    time.Duration
}

type WorkerConfig struct {
	RefreshInterval time.Duration
	BatchSize       int
	Enabled         bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			HTTPPort: getEnv("HTTP_PORT", "8080"),
			GRPCPort: getEnv("GRPC_PORT", "50051"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "market_data"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      getEnvAsDuration("REDIS_TTL", 60*time.Second),
		},
		CoinGecko: CoinGeckoConfig{
			APIKey:    getEnv("COINGECKO_API_KEY", ""),
			BaseURL:   getEnv("COINGECKO_BASE_URL", "https://api.coingecko.com/api/v3"),
			RateLimit: getEnvAsInt("COINGECKO_RATE_LIMIT", 50),
			Timeout:   getEnvAsDuration("COINGECKO_TIMEOUT", 10*time.Second),
		},
		Worker: WorkerConfig{
			RefreshInterval: getEnvAsDuration("WORKER_REFRESH_INTERVAL", 30*time.Second),
			BatchSize:       getEnvAsInt("WORKER_BATCH_SIZE", 50),
			Enabled:         getEnvAsBool("WORKER_ENABLED", true),
		},
	}

	return cfg, nil
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
