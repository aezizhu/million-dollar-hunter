package transformer

import (
	"testing"
)

func TestTransformAlchemyTransfers(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantLen int
		wantErr bool
	}{
		{
			name:    "nil data",
			data:    nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "empty map",
			data:    map[string]interface{}{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "valid alchemy response",
			data: map[string]interface{}{
				"transfers": []interface{}{
					map[string]interface{}{
						"hash":     "0x123",
						"from":     "0xabc",
						"to":       "0xdef",
						"category": "erc20",
						"asset":    "USDC",
						"rawContract": map[string]interface{}{
							"value":   "1000000",
							"address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
						},
						"metadata": map[string]interface{}{
							"blockTimestamp": "2024-01-01T00:00:00Z",
						},
					},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "skip transfer with empty hash",
			data: map[string]interface{}{
				"transfers": []interface{}{
					map[string]interface{}{
						"hash": "",
						"from": "0xabc",
						"to":   "0xdef",
					},
					map[string]interface{}{
						"hash": "0x456",
						"from": "0xghi",
						"to":   "0xjkl",
					},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "multiple transfers",
			data: map[string]interface{}{
				"transfers": []interface{}{
					map[string]interface{}{
						"hash": "0x123",
						"from": "0xabc",
						"to":   "0xdef",
					},
					map[string]interface{}{
						"hash": "0x456",
						"from": "0xghi",
						"to":   "0xjkl",
					},
				},
			},
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformAlchemyTransfers(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformAlchemyTransfers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("TransformAlchemyTransfers() got %d transactions, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestTransformAlchemyTransfers_Fields(t *testing.T) {
	data := map[string]interface{}{
		"transfers": []interface{}{
			map[string]interface{}{
				"hash":     "0xabc123",
				"from":     "0x111",
				"to":       "0x222",
				"category": "erc20",
				"asset":    "USDT",
				"rawContract": map[string]interface{}{
					"value":   "5000000",
					"address": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				},
				"metadata": map[string]interface{}{
					"blockTimestamp": "2024-03-15T10:30:00Z",
				},
			},
		},
	}

	txs, err := TransformAlchemyTransfers(data)
	if err != nil {
		t.Fatalf("TransformAlchemyTransfers() error = %v", err)
	}

	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	tx := txs[0]
	if tx.Hash != "0xabc123" {
		t.Errorf("Hash = %v, want %v", tx.Hash, "0xabc123")
	}
	if tx.From != "0x111" {
		t.Errorf("From = %v, want %v", tx.From, "0x111")
	}
	if tx.To != "0x222" {
		t.Errorf("To = %v, want %v", tx.To, "0x222")
	}
	if tx.Amount != "5000000" {
		t.Errorf("Amount = %v, want %v", tx.Amount, "5000000")
	}
	if tx.Symbol != "USDT" {
		t.Errorf("Symbol = %v, want %v", tx.Symbol, "USDT")
	}
	if tx.TokenAddress != "0xdAC17F958D2ee523a2206206994597C13D831ec7" {
		t.Errorf("TokenAddress = %v, want %v", tx.TokenAddress, "0xdAC17F958D2ee523a2206206994597C13D831ec7")
	}
	if tx.Type != "erc20" {
		t.Errorf("Type = %v, want %v", tx.Type, "erc20")
	}
}
