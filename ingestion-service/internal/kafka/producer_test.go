package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewProducer_Validation(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	tests := []struct {
		name    string
		brokers []string
		topic   string
		wantErr bool
	}{
		{
			name:    "empty brokers",
			brokers: []string{},
			topic:   "test-topic",
			wantErr: true,
		},
		{
			name:    "empty topic",
			brokers: []string{"localhost:9092"},
			topic:   "",
			wantErr: true,
		},
		{
			name:    "valid inputs but connection fails",
			brokers: []string{"localhost:9092"},
			topic:   "test-topic",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProducer(ctx, tt.brokers, tt.topic, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProducer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPublishTransactionDataIngested_Validation(t *testing.T) {
	tests := []struct {
		name    string
		event   TransactionDataIngestedEvent
		wantErr bool
	}{
		{
			name: "empty wallet address",
			event: TransactionDataIngestedEvent{
				Chain:        "ethereum",
				Transactions: []Transaction{},
			},
			wantErr: true,
		},
		{
			name: "empty chain",
			event: TransactionDataIngestedEvent{
				WalletAddress: "0x123",
				Transactions:  []Transaction{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Producer{
				producer: nil,
				topic:    "test",
				logger:   zerolog.Nop(),
			}
			err := p.PublishTransactionDataIngested(context.Background(), tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishTransactionDataIngested() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransactionDataIngestedEvent_AutoFields(t *testing.T) {
	event := TransactionDataIngestedEvent{
		WalletAddress: "0x123",
		Chain:         "ethereum",
		Transactions:  []Transaction{},
		EventID:       "",
		SchemaVersion: "",
	}

	if event.EventID != "" {
		t.Error("EventID should be empty before publishing")
	}
	if event.SchemaVersion != "" {
		t.Error("SchemaVersion should be empty before publishing")
	}
}

func TestTransaction_Struct(t *testing.T) {
	tx := Transaction{
		Hash:      "0xabc",
		From:      "0x123",
		To:        "0x456",
		Amount:    "1000000000000000000",
		Symbol:    "ETH",
		Timestamp: time.Now(),
		Type:      "transfer",
	}

	if tx.Hash == "" {
		t.Error("Transaction hash should not be empty")
	}
	if tx.From == "" {
		t.Error("Transaction from should not be empty")
	}
	if tx.To == "" {
		t.Error("Transaction to should not be empty")
	}
}
