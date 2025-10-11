package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
	logger   zerolog.Logger
}

type TransactionDataIngestedEvent struct {
	EventID           string `json:"event_id"`
	Timestamp         string `json:"timestamp"`
	WalletAddress     string `json:"wallet_address"`
	Chain             string `json:"chain"`
	DataSource        string `json:"data_source"`
	TransactionCount  int    `json:"transaction_count"`
	IngestionJobID    string `json:"ingestion_job_id"`
	Status            string `json:"status"`
}

func NewProducer(brokers []string, topic string, logger zerolog.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}, nil
}

func (p *Producer) PublishTransactionDataIngested(ctx context.Context, event TransactionDataIngestedEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if event.IngestionJobID == "" {
		event.IngestionJobID = uuid.New().String()
	}

	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Error().Err(err).Msg("marshal kafka event")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.WalletAddress),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error().Err(err).Str("topic", p.topic).Msg("send kafka message")
		return err
	}

	p.logger.Debug().
		Str("event_id", event.EventID).
		Str("wallet", event.WalletAddress).
		Str("chain", event.Chain).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("published transaction data ingested event")

	return nil
}

func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
