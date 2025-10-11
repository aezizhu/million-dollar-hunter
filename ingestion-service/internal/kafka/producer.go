package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/metrics"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
	logger   zerolog.Logger
	metrics  *metrics.KafkaMetrics
}

type Transaction struct {
	Hash         string    `json:"hash"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	Amount       string    `json:"amount"`
	Symbol       string    `json:"symbol"`
	TokenAddress string    `json:"token_address"`
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`
}

type TransactionDataIngestedEvent struct {
	SchemaVersion string        `json:"schema_version"`
	EventID       string        `json:"event_id"`
	WalletAddress string        `json:"wallet_address"`
	Chain         string        `json:"chain"`
	Transactions  []Transaction `json:"transactions"`
}

func NewProducer(ctx context.Context, brokers []string, topic string, logger zerolog.Logger, metrics *metrics.KafkaMetrics) (*Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("at least one broker required")
	}
	if topic == "" {
		return nil, fmt.Errorf("topic required")
	}

	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 3
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Idempotent = true
	config.Producer.MaxMessageBytes = 1000000

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		if metrics != nil {
			metrics.ConnectionErrors.Inc()
			metrics.Connected.Set(0)
		}
		return nil, fmt.Errorf("create producer: %w", err)
	}

	if metrics != nil {
		metrics.Connected.Set(1)
	}

	logger.Info().
		Strs("brokers", brokers).
		Str("topic", topic).
		Msg("kafka producer initialized")

	return &Producer{
		producer: producer,
		topic:    topic,
		logger:   logger,
		metrics:  metrics,
	}, nil
}

func (p *Producer) PublishTransactionDataIngested(ctx context.Context, event TransactionDataIngestedEvent) error {
	if event.WalletAddress == "" {
		return fmt.Errorf("wallet_address is required")
	}
	if event.Chain == "" {
		return fmt.Errorf("chain is required")
	}

	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.SchemaVersion == "" {
		event.SchemaVersion = "1.0.0"
	}

	start := time.Now()
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Error().Err(err).Msg("marshal kafka event failed")
		if p.metrics != nil {
			p.metrics.ObservePublish(p.topic, err, 0, start)
		}
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.WalletAddress),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event_id"), Value: []byte(event.EventID)},
			{Key: []byte("schema_version"), Value: []byte(event.SchemaVersion)},
		},
	}

	partition, offset, err := p.producer.SendMessage(msg)
	
	if p.metrics != nil {
		p.metrics.ObservePublish(p.topic, err, len(data), start)
	}
	
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("topic", p.topic).
			Str("wallet", event.WalletAddress).
			Msg("send kafka message failed")
		return fmt.Errorf("send message: %w", err)
	}

	duration := time.Since(start)
	p.logger.Info().
		Str("event_id", event.EventID).
		Str("wallet", event.WalletAddress).
		Str("chain", event.Chain).
		Int("transaction_count", len(event.Transactions)).
		Int32("partition", partition).
		Int64("offset", offset).
		Dur("duration_ms", duration).
		Msg("published transaction data ingested event")

	return nil
}

func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
