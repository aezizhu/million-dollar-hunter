package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTransactionDataIngestedEvent_MarshalUnmarshal(t *testing.T) {
	original := TransactionDataIngestedEvent{
		SchemaVersion: "1.0.0",
		EventID:       "test-event-123",
		WalletAddress: "0xabc123",
		Chain:         "ethereum",
		Transactions: []Transaction{
			{
				Hash:         "0x789abc",
				From:         "0x111",
				To:           "0x222",
				Amount:       "1000000",
				Symbol:       "USDC",
				TokenAddress: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
				Timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Type:         "erc20",
			},
			{
				Hash:         "0xdef456",
				From:         "0x333",
				To:           "0x444",
				Amount:       "5000000",
				Symbol:       "USDT",
				TokenAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				Timestamp:    time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
				Type:         "erc20",
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TransactionDataIngestedEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion = %v, want %v", decoded.SchemaVersion, original.SchemaVersion)
	}
	if decoded.EventID != original.EventID {
		t.Errorf("EventID = %v, want %v", decoded.EventID, original.EventID)
	}
	if decoded.WalletAddress != original.WalletAddress {
		t.Errorf("WalletAddress = %v, want %v", decoded.WalletAddress, original.WalletAddress)
	}
	if decoded.Chain != original.Chain {
		t.Errorf("Chain = %v, want %v", decoded.Chain, original.Chain)
	}
	if len(decoded.Transactions) != len(original.Transactions) {
		t.Fatalf("Transactions length = %d, want %d", len(decoded.Transactions), len(original.Transactions))
	}

	for i, tx := range decoded.Transactions {
		orig := original.Transactions[i]
		if tx.Hash != orig.Hash {
			t.Errorf("Transaction[%d].Hash = %v, want %v", i, tx.Hash, orig.Hash)
		}
		if tx.From != orig.From {
			t.Errorf("Transaction[%d].From = %v, want %v", i, tx.From, orig.From)
		}
		if tx.To != orig.To {
			t.Errorf("Transaction[%d].To = %v, want %v", i, tx.To, orig.To)
		}
		if tx.Amount != orig.Amount {
			t.Errorf("Transaction[%d].Amount = %v, want %v", i, tx.Amount, orig.Amount)
		}
		if tx.Symbol != orig.Symbol {
			t.Errorf("Transaction[%d].Symbol = %v, want %v", i, tx.Symbol, orig.Symbol)
		}
		if tx.TokenAddress != orig.TokenAddress {
			t.Errorf("Transaction[%d].TokenAddress = %v, want %v", i, tx.TokenAddress, orig.TokenAddress)
		}
		if !tx.Timestamp.Equal(orig.Timestamp) {
			t.Errorf("Transaction[%d].Timestamp = %v, want %v", i, tx.Timestamp, orig.Timestamp)
		}
		if tx.Type != orig.Type {
			t.Errorf("Transaction[%d].Type = %v, want %v", i, tx.Type, orig.Type)
		}
	}
}

func TestPortfolioServiceCompatibility(t *testing.T) {
	type PortfolioTransaction struct {
		Hash         string    `json:"hash"`
		From         string    `json:"from"`
		To           string    `json:"to"`
		Amount       string    `json:"amount"`
		Symbol       string    `json:"symbol"`
		TokenAddress string    `json:"token_address"`
		Timestamp    time.Time `json:"timestamp"`
		Type         string    `json:"type"`
	}

	type PortfolioEvent struct {
		WalletAddress string                 `json:"wallet_address"`
		Chain         string                 `json:"chain"`
		Transactions  []PortfolioTransaction `json:"transactions"`
	}

	producerEvent := TransactionDataIngestedEvent{
		SchemaVersion: "1.0.0",
		EventID:       "test-123",
		WalletAddress: "0xabc",
		Chain:         "ethereum",
		Transactions: []Transaction{
			{
				Hash:         "0x123",
				From:         "0x111",
				To:           "0x222",
				Amount:       "1000",
				Symbol:       "USDC",
				TokenAddress: "0xA0b",
				Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Type:         "erc20",
			},
		},
	}

	data, err := json.Marshal(producerEvent)
	if err != nil {
		t.Fatalf("Marshal producer event failed: %v", err)
	}

	var consumerEvent PortfolioEvent
	if err := json.Unmarshal(data, &consumerEvent); err != nil {
		t.Fatalf("Unmarshal to portfolio event failed: %v", err)
	}

	if consumerEvent.WalletAddress != producerEvent.WalletAddress {
		t.Errorf("WalletAddress mismatch")
	}
	if consumerEvent.Chain != producerEvent.Chain {
		t.Errorf("Chain mismatch")
	}
	if len(consumerEvent.Transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(consumerEvent.Transactions))
	}

	tx := consumerEvent.Transactions[0]
	origTx := producerEvent.Transactions[0]
	
	if tx.Hash != origTx.Hash {
		t.Errorf("Hash mismatch: got %v, want %v", tx.Hash, origTx.Hash)
	}
	if tx.TokenAddress != origTx.TokenAddress {
		t.Errorf("TokenAddress mismatch: got %v, want %v", tx.TokenAddress, origTx.TokenAddress)
	}
	if tx.Symbol != origTx.Symbol {
		t.Errorf("Symbol mismatch: got %v, want %v", tx.Symbol, origTx.Symbol)
	}
}
