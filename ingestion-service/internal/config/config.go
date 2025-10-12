package config

import (
	"fmt"
	"os"
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

func Load() (*Config, error) {
	_ = os.Setenv("TZ", "UTC")
	kafkaBrokers := get("KAFKA_BROKERS", "localhost:9092")
	return &Config{
		DatabaseURL:             get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable"),
		RedisAddr:               get("REDIS_ADDR", "localhost:6379"),
		AlchemyBaseURL:          get("ALCHEMY_BASE_URL", "http://localhost:8080/alchemy"),
		AlchemyAPIKey:           get("ALCHEMY_API_KEY", "test"),
		MoralisBaseURL:          get("MORALIS_BASE_URL", "http://localhost:8080/moralis"),
		MoralisAPIKey:           get("MORALIS_API_KEY", "test"),
		RateLimitAlchemyRPS:     get("RATE_LIMIT_ALCHEMY_RPS", "20"),
		RateLimitMoralisRPS:     get("RATE_LIMIT_MORALIS_RPS", "10"),
		UseAPIMocks:             get("USE_API_MOCKS", "true"),
		HTTPPort:                get("HTTP_PORT", "8090"),
		JobQueueSize:            getInt("JOB_QUEUE_SIZE", 64),
		KafkaBrokers:            kafkaBrokers,
		KafkaTopicTxIngested:    get("KAFKA_TOPIC_TX_INGESTED", "TransactionDataIngested"),
		KafkaTopicWalletTracking: get("KAFKA_TOPIC_WALLET_TRACKING", "WalletTrackingRequested"),
		KafkaConsumerGroupID:    get("KAFKA_CONSUMER_GROUP_ID", "ingestion-service"),
		KafkaEnabled:            kafkaBrokers != "",
	}, nil
}
