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
	ExportCleanupTTL             time.Duration `env:"EXPORT_CLEANUP_TTL" envDefault:"1h"`
	ExportCleanupInterval        time.Duration `env:"EXPORT_CLEANUP_INTERVAL" envDefault:"15m"`
	MarketDataServiceAddr        string        `env:"MARKET_DATA_SERVICE_ADDR,notEmpty" envDefault:"localhost:50051"`
	MarketDataSingleTimeout      time.Duration `env:"MARKET_DATA_SINGLE_TIMEOUT" envDefault:"5s"`
	MarketDataBatchTimeout       time.Duration `env:"MARKET_DATA_BATCH_TIMEOUT" envDefault:"10s"`
	MarketDataTLSEnabled         bool          `env:"MARKET_DATA_TLS_ENABLED" envDefault:"false"`
	MarketDataCAFile             string        `env:"MARKET_DATA_CA_FILE" envDefault:""`
	MarketDataServerName         string        `env:"MARKET_DATA_SERVER_NAME" envDefault:""`
	// Logging configuration - reserved for future structured logging implementation
	LogLevel                     string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat                    string        `env:"LOG_FORMAT" envDefault:"json"`
	// Observability configuration
	OTELEndpoint                 string        `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	PrometheusNamespace          string        `env:"PROMETHEUS_NAMESPACE" envDefault:"portfolio_service"`
	MetricsAddr                  string        `env:"METRICS_ADDR" envDefault:""` // e.g., :9090 to enable /metrics
}
