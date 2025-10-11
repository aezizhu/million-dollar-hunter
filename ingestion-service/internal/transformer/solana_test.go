package transformer

import (
	"testing"
)

func TestTransformSolanaTransactions(t *testing.T) {
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
			name:    "empty array",
			data:    []interface{}{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "valid solana transaction",
			data: []interface{}{
				map[string]interface{}{
					"signature": "5j7s6NiJS3JAkvgkoc18WVAsiSaci2pxB2A6ueCJP4tprA2TFg9wSyTLeYouxPBJEMzJinENTkpA52YStRW5Dia7",
					"blockTime": float64(1640000000),
					"meta": map[string]interface{}{
						"preBalances":  []interface{}{float64(1000000000), float64(5000000000)},
						"postBalances": []interface{}{float64(900000000), float64(5100000000)},
					},
					"transaction": map[string]interface{}{
						"message": map[string]interface{}{
							"accountKeys": []interface{}{
								"7UX2i7SucgLMQcfZ75s3VXmZZY4YRUyJN9X1RgfMoDUi",
								"FNNvb1AFDnDVPkocEri8mWbJ1952HQZtFLuwPiUjSJQ",
							},
						},
					},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "transaction with empty signature",
			data: []interface{}{
				map[string]interface{}{
					"signature": "",
					"blockTime": float64(1640000000),
				},
				map[string]interface{}{
					"signature": "validhash123",
					"blockTime": float64(1640000000),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "transaction with hash fallback",
			data: []interface{}{
				map[string]interface{}{
					"hash":      "fallbackhash456",
					"blockTime": float64(1640000000),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformSolanaTransactions(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformSolanaTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("TransformSolanaTransactions() got %d transactions, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestTransformSolanaTransactions_Fields(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{
			"signature": "5j7s6NiJS3JAkvgkoc18WVAsiSaci2pxB2A6ueCJP4tprA2TFg9wSyTLeYouxPBJEMzJinENTkpA52YStRW5Dia7",
			"blockTime": float64(1640000000),
			"meta": map[string]interface{}{
				"preBalances":  []interface{}{float64(1000000000), float64(5000000000)},
				"postBalances": []interface{}{float64(900000000), float64(5100000000)},
			},
			"transaction": map[string]interface{}{
				"message": map[string]interface{}{
					"accountKeys": []interface{}{
						"7UX2i7SucgLMQcfZ75s3VXmZZY4YRUyJN9X1RgfMoDUi",
						"FNNvb1AFDnDVPkocEri8mWbJ1952HQZtFLuwPiUjSJQ",
					},
				},
			},
		},
	}

	txs, err := TransformSolanaTransactions(data)
	if err != nil {
		t.Fatalf("TransformSolanaTransactions() error = %v", err)
	}

	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	tx := txs[0]
	
	expectedSig := "5j7s6NiJS3JAkvgkoc18WVAsiSaci2pxB2A6ueCJP4tprA2TFg9wSyTLeYouxPBJEMzJinENTkpA52YStRW5Dia7"
	if tx.Hash != expectedSig {
		t.Errorf("Hash = %v, want %v", tx.Hash, expectedSig)
	}

	if tx.From != "7UX2i7SucgLMQcfZ75s3VXmZZY4YRUyJN9X1RgfMoDUi" {
		t.Errorf("From = %v, want 7UX2i7SucgLMQcfZ75s3VXmZZY4YRUyJN9X1RgfMoDUi", tx.From)
	}

	if tx.To != "FNNvb1AFDnDVPkocEri8mWbJ1952HQZtFLuwPiUjSJQ" {
		t.Errorf("To = %v, want FNNvb1AFDnDVPkocEri8mWbJ1952HQZtFLuwPiUjSJQ", tx.To)
	}

	if tx.Amount != "-100000000" {
		t.Errorf("Amount = %v, want -100000000", tx.Amount)
	}

	if tx.Symbol != "SOL" {
		t.Errorf("Symbol = %v, want SOL", tx.Symbol)
	}

	expectedTokenAddr := "So11111111111111111111111111111111111111112"
	if tx.TokenAddress != expectedTokenAddr {
		t.Errorf("TokenAddress = %v, want %v (Solana native mint)", tx.TokenAddress, expectedTokenAddr)
	}

	if tx.Type != "transfer" {
		t.Errorf("Type = %v, want transfer", tx.Type)
	}

	if tx.Timestamp.Unix() != 1640000000 {
		t.Errorf("Timestamp = %v, want 1640000000", tx.Timestamp.Unix())
	}
}

func TestTransformSolanaTransactions_TimestampHandling(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{
			"signature": "testHash",
		},
	}

	txs, err := TransformSolanaTransactions(data)
	if err != nil {
		t.Fatalf("TransformSolanaTransactions() error = %v", err)
	}

	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	tx := txs[0]
	
	if tx.Timestamp.IsZero() {
		t.Error("Timestamp should be set to current time when blockTime is missing")
	}
}
