package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/secrets"
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

type dbSecret struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"sslmode"`
}

type redisSecret struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func loadSecrets(ctx context.Context) secrets.Client {
	if os.Getenv("SECRETS_PROVIDER") == "aws" {
		c, err := secrets.NewAWS(ctx, secrets.AWSConfig{
			Config: secrets.Config{
				CacheTTL:        time.Hour,
				RefreshInterval: time.Minute,
			},
			Region: os.Getenv("AWS_REGION"),
		})
		if err == nil {
			return c
		}
	}
	return secrets.NewEnv(secrets.Config{
		CacheTTL:        time.Hour,
		RefreshInterval: time.Minute,
	})
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sc := loadSecrets(ctx)

	var ds dbSecret
	if err := sc.GetJSON(ctx, "market/db", &ds); err == nil && ds.Host != "" {
		cfg.Database.Host = ds.Host
		if ds.Port != 0 {
			cfg.Database.Port = ds.Port
		}
		if ds.User != "" {
			cfg.Database.User = ds.User
		}
		if ds.Password != "" {
			cfg.Database.Password = ds.Password
		}
		if ds.Name != "" {
			cfg.Database.DBName = ds.Name
		}
		if ds.SSLMode != "" {
			cfg.Database.SSLMode = ds.SSLMode
		}
	}

	var rs redisSecret
	if err := sc.GetJSON(ctx, "shared/redis", &rs); err == nil {
		if rs.Addr != "" {
			host := rs.Addr
			port := 0
			for i := len(host) - 1; i >= 0; i-- {
				if host[i] == ':' {
					if p, err := strconv.Atoi(host[i+1:]); err == nil {
						port = p
					}
					host = host[:i]
					break
				}
			}
			if host != "" {
				cfg.Redis.Host = host
			}
			if port > 0 {
				cfg.Redis.Port = port
			}
		}
		if rs.Password != "" {
			cfg.Redis.Password = rs.Password
		}
		if rs.DB != 0 {
			cfg.Redis.DB = rs.DB
		}
	}

	if s, err := sc.Get(ctx, "market/coingecko"); err == nil && s != "" {
		cfg.CoinGecko.APIKey = s
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
