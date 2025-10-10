package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPPort string
	
	JWTSecret          string
	JWTExpiryMinutes   int
	
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	
	RateLimitPerMinute int
	RateLimitBurst     int
	
	AuthServiceAddr      string
	PortfolioServiceAddr string
	MarketDataServiceAddr string
	
	LogLevel string
	
	MVPUsername string
	MVPPassword string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:              getEnv("HTTP_PORT", "8080"),
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiryMinutes:      getEnvAsInt("JWT_EXPIRY_MINUTES", 15),
		RedisAddr:             getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		RedisDB:               getEnvAsInt("REDIS_DB", 0),
		RateLimitPerMinute:    getEnvAsInt("RATE_LIMIT_PER_MINUTE", 100),
		RateLimitBurst:        getEnvAsInt("RATE_LIMIT_BURST", 200),
		AuthServiceAddr:       getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		PortfolioServiceAddr:  getEnv("PORTFOLIO_SERVICE_ADDR", "localhost:50052"),
		MarketDataServiceAddr: getEnv("MARKET_DATA_SERVICE_ADDR", "localhost:50053"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		MVPUsername:           getEnv("MVP_USERNAME", "aezi"),
		MVPPassword:           getEnv("MVP_PASSWORD", "Aa@123456789"),
	}
	
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET cannot be empty")
	}
	if c.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR cannot be empty")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
