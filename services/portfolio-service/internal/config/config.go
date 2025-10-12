package config

import "time"

type Config struct {
	GRPCAddr              string        `env:"GRPC_ADDR,notEmpty" envDefault:":8081"`
	DatabaseURL           string        `env:"DATABASE_URL,notEmpty" envDefault:"postgres://portfolio:portfolio@localhost:5432/portfolio?sslmode=disable"`
	DBMaxConns            int32         `env:"DB_MAX_CONNS" envDefault:"20"`
	DBMinConns            int32         `env:"DB_MIN_CONNS" envDefault:"5"`
	DBMaxConnLifetime     time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"1h"`
	DBMaxConnIdleTime     time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"30m"`
	DBHealthCheckPeriod   time.Duration `env:"DB_HEALTH_CHECK_PERIOD" envDefault:"1m"`
	KafkaBrokers          string        `env:"KAFKA_BROKERS,notEmpty" envDefault:"localhost:9092"`
	TopicTxIngested       string        `env:"TOPIC_TRANSACTION_INGESTED,notEmpty" envDefault:"TransactionDataIngested"`
	TopicPortfolioUpdated string        `env:"TOPIC_PORTFOLIO_UPDATED,notEmpty" envDefault:"PortfolioUpdated"`
	GroupID               string        `env:"KAFKA_GROUP_ID,notEmpty" envDefault:"portfolio-service"`
	ExportDir             string        `env:"EXPORT_DIR,notEmpty" envDefault:"/data/exports"`
	ExportCleanupTTL      time.Duration `env:"EXPORT_CLEANUP_TTL" envDefault:"1h"`
	ExportCleanupInterval time.Duration `env:"EXPORT_CLEANUP_INTERVAL" envDefault:"15m"`
	// Logging configuration - reserved for future structured logging implementation
	LogLevel              string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat             string        `env:"LOG_FORMAT" envDefault:"json"`
	// Observability configuration - reserved for future instrumentation
	OTELEndpoint          string        `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	PrometheusNamespace   string        `env:"PROMETHEUS_NAMESPACE" envDefault:"portfolio_service"`

	MarketDataGRPCAddr   string        `env:"MARKET_DATA_GRPC_ADDR" envDefault:"localhost:50051"`
	MarketDataTimeout    time.Duration `env:"MARKET_DATA_TIMEOUT" envDefault:"2s"`
	MarketDataTLSEnabled bool          `env:"MARKET_DATA_TLS_ENABLED" envDefault:"false"`
}
