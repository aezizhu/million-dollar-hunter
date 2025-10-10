package config

type Config struct {
	GRPCAddr            string `env:"GRPC_ADDR,notEmpty" envDefault:":8081"`
	DatabaseURL         string `env:"DATABASE_URL,notEmpty" envDefault:"postgres://portfolio:portfolio@localhost:5432/portfolio?sslmode=disable"`
	KafkaBrokers        string `env:"KAFKA_BROKERS,notEmpty" envDefault:"localhost:9092"`
	TopicTxIngested     string `env:"TOPIC_TRANSACTION_INGESTED,notEmpty" envDefault:"TransactionDataIngested"`
	TopicPortfolioUpdated string `env:"TOPIC_PORTFOLIO_UPDATED,notEmpty" envDefault:"PortfolioUpdated"`
	GroupID             string `env:"KAFKA_GROUP_ID,notEmpty" envDefault:"portfolio-service"`
	ExportDir           string `env:"EXPORT_DIR,notEmpty" envDefault:"/data/exports"`
}
