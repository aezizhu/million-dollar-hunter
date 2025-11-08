// Package config provides configuration management for the ingestion service.
// Copyright (c) 2025 aezizhu. All rights reserved.
// Author: aezizhu
// Repository: github.com/aezizhu/million-dollar-hunter
package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aezizhu/million-dollar-hunter/pkg/secrets"
)

type Config struct {
	DatabaseURL string
	RedisAddr   string

	AlchemyBaseURL string
	AlchemyAPIKey  string
	MoralisBaseURL string
	MoralisAPIKey  string

	RateLimitAlchemyRPS string
	RateLimitMoralisRPS string

	UseAPIMocks string
	HTTPPort    string

	JobQueueSize int

	KafkaBrokers                string
	KafkaTopicTxIngested        string
	KafkaTopicWalletTracking    string
	KafkaConsumerGroupID        string
	KafkaEnabled                bool
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return def
}

type redisSecret struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type apisSecret struct {
	AlchemyAPIKey string `json:"alchemy_api_key"`
	MoralisAPIKey string `json:"moralis_api_key"`
}

func loadSecrets(ctx context.Context) secrets.Client {
	provider := os.Getenv("SECRETS_PROVIDER")
	if provider == "aws" {
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
	_ = os.Setenv("TZ", "UTC")
	kafkaBrokers := get("KAFKA_BROKERS", "localhost:9092")

	cfg := &Config{
		DatabaseURL:              get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable"),
		RedisAddr:                get("REDIS_ADDR", "localhost:6379"),
		AlchemyBaseURL:           get("ALCHEMY_BASE_URL", "http://localhost:8080/alchemy"),
		AlchemyAPIKey:            get("ALCHEMY_API_KEY", "test"),
		MoralisBaseURL:           get("MORALIS_BASE_URL", "http://localhost:8080/moralis"),
		MoralisAPIKey:            get("MORALIS_API_KEY", "test"),
		RateLimitAlchemyRPS:      get("RATE_LIMIT_ALCHEMY_RPS", "20"),
		RateLimitMoralisRPS:      get("RATE_LIMIT_MORALIS_RPS", "10"),
		UseAPIMocks:              get("USE_API_MOCKS", "true"),
		HTTPPort:                 get("HTTP_PORT", "8090"),
		JobQueueSize:             getInt("JOB_QUEUE_SIZE", 64),
		KafkaBrokers:             kafkaBrokers,
		KafkaTopicTxIngested:     get("KAFKA_TOPIC_TX_INGESTED", "TransactionDataIngested"),
		KafkaTopicWalletTracking: get("KAFKA_TOPIC_WALLET_TRACKING", "WalletTrackingRequested"),
		KafkaConsumerGroupID:     get("KAFKA_CONSUMER_GROUP_ID", "ingestion-service"),
		KafkaEnabled:             kafkaBrokers != "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sc := loadSecrets(ctx)

	if s, err := sc.Get(ctx, "ingestion/db_url"); err == nil && s != "" {
		cfg.DatabaseURL = s
	}
	var rs redisSecret
	if err := sc.GetJSON(ctx, "shared/redis", &rs); err == nil && rs.Addr != "" {
		cfg.RedisAddr = rs.Addr
	}
	var as apisSecret
	if err := sc.GetJSON(ctx, "ingestion/apis", &as); err == nil {
		if as.AlchemyAPIKey != "" {
			cfg.AlchemyAPIKey = as.AlchemyAPIKey
		}
		if as.MoralisAPIKey != "" {
			cfg.MoralisAPIKey = as.MoralisAPIKey
		}
	}

	return cfg, nil
}
