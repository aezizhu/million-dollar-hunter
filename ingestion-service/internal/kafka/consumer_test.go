package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

type mockHandler struct {
	handleFunc func(ctx context.Context, event WalletTrackingRequestedEvent) error
}

func (m *mockHandler) HandleWalletTrackingRequest(ctx context.Context, event WalletTrackingRequestedEvent) error {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, event)
	}
	return nil
}

type mockMetrics struct {
	consumeCalls    int
	rebalanceCalls  int
	lastTopic       string
	lastErr         error
}

func (m *mockMetrics) ObserveConsume(topic string, err error, start time.Time) {
	m.consumeCalls++
	m.lastTopic = topic
	m.lastErr = err
}

func (m *mockMetrics) ObserveRebalance() {
	m.rebalanceCalls++
}

func TestNewConsumer_Validation(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()
	handler := &mockHandler{}
	metrics := &mockMetrics{}

	tests := []struct {
		name        string
		brokers     []string
		groupID     string
		topic       string
		handler     WalletRequestHandler
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty_brokers",
			brokers:     []string{},
			groupID:     "test-group",
			topic:       "test-topic",
			handler:     handler,
			wantErr:     true,
			errContains: "at least one broker required",
		},
		{
			name:        "empty_groupID",
			brokers:     []string{"localhost:9092"},
			groupID:     "",
			topic:       "test-topic",
			handler:     handler,
			wantErr:     true,
			errContains: "groupID required",
		},
		{
			name:        "empty_topic",
			brokers:     []string{"localhost:9092"},
			groupID:     "test-group",
			topic:       "",
			handler:     handler,
			wantErr:     true,
			errContains: "topic required",
		},
		{
			name:        "nil_handler",
			brokers:     []string{"localhost:9092"},
			groupID:     "test-group",
			topic:       "test-topic",
			handler:     nil,
			wantErr:     true,
			errContains: "handler required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConsumer(ctx, tt.brokers, tt.groupID, tt.topic, logger, tt.handler, metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConsumer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewConsumer() error = %v, want error containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestWalletTrackingRequestedEvent_MarshalUnmarshal(t *testing.T) {
	original := WalletTrackingRequestedEvent{
		SchemaVersion: "1.0.0",
		EventID:       "test-event-123",
		Timestamp:     time.Now().UTC(),
		UserID:        "user-456",
		WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		Chain:         "ethereum",
		Nickname:      "My Wallet",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded WalletTrackingRequestedEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
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
}

func TestValidateEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       *WalletTrackingRequestedEvent
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_ethereum",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Chain:         "ethereum",
			},
			wantErr: false,
		},
		{
			name: "valid_solana",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",
				Chain:         "solana",
			},
			wantErr: false,
		},
		{
			name: "empty_wallet_address",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "",
				Chain:         "ethereum",
			},
			wantErr:     true,
			errContains: "wallet_address",
		},
		{
			name: "empty_chain",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Chain:         "",
			},
			wantErr:     true,
			errContains: "chain",
		},
		{
			name: "invalid_chain",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Chain:         "bitcoin",
			},
			wantErr:     true,
			errContains: "invalid chain",
		},
		{
			name: "invalid_ethereum_address_too_short",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "0x123",
				Chain:         "ethereum",
			},
			wantErr:     true,
			errContains: "invalid EVM address",
		},
		{
			name: "invalid_ethereum_address_no_0x",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Chain:         "ethereum",
			},
			wantErr:     true,
			errContains: "invalid EVM address",
		},
		{
			name: "invalid_solana_address",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "invalid",
				Chain:         "solana",
			},
			wantErr:     true,
			errContains: "invalid Solana address",
		},
		{
			name: "uppercase_chain_rejected",
			event: &WalletTrackingRequestedEvent{
				WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Chain:         "ETHEREUM",
			},
			wantErr:     true,
			errContains: "invalid chain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEvent(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("validateEvent() error = %v, want error containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidateWalletAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		chain   string
		wantErr bool
	}{
		{"valid_ethereum", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "ethereum", false},
		{"valid_bsc", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "bsc", false},
		{"valid_polygon", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "polygon", false},
		{"valid_arbitrum", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "arbitrum", false},
		{"valid_optimism", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "optimism", false},
		{"valid_solana", "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK", "solana", false},
		{"invalid_evm_too_short", "0x123", "ethereum", true},
		{"invalid_evm_no_prefix", "742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "ethereum", true},
		{"invalid_solana_too_short", "abc", "solana", true},
		{"invalid_solana_invalid_chars", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "solana", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWalletAddress(tt.address, tt.chain)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWalletAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsPermanentError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"validation_error", &ValidationError{Field: "test", Message: "test"}, true},
		{"json_syntax_error", &json.SyntaxError{}, true},
		{"unmarshal_error", errors.New("unmarshal failed"), true},
		{"malformed_error", errors.New("malformed data"), true},
		{"connection_error", errors.New("connection refused"), false},
		{"timeout_error", errors.New("request timeout"), false},
		{"unknown_error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPermanentError(tt.err); got != tt.want {
				t.Errorf("isPermanentError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"transient_error", &TransientError{Underlying: errors.New("test")}, true},
		{"connection_error", errors.New("connection refused"), true},
		{"timeout_error", errors.New("request timeout"), true},
		{"temporary_error", errors.New("temporary failure"), true},
		{"unavailable_error", errors.New("service unavailable"), true},
		{"deadline_error", errors.New("context deadline exceeded"), true},
		{"queue_full_error", errors.New("queue full"), true},
		{"validation_error", &ValidationError{Field: "test", Message: "test"}, false},
		{"unmarshal_error", errors.New("unmarshal failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransientError(tt.err); got != tt.want {
				t.Errorf("isTransientError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "wallet_address", Message: "required field is empty"}
	expected := "validation error for field 'wallet_address': required field is empty"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Error("ValidationError should be detectable with errors.As")
	}
}

func TestTransientError(t *testing.T) {
	underlying := errors.New("connection failed")
	err := &TransientError{Underlying: underlying}
	
	if !contains(err.Error(), "transient error") {
		t.Errorf("TransientError.Error() = %v, want to contain 'transient error'", err.Error())
	}

	if errors.Unwrap(err) != underlying {
		t.Error("TransientError.Unwrap() should return underlying error")
	}

	var transErr *TransientError
	if !errors.As(err, &transErr) {
		t.Error("TransientError should be detectable with errors.As")
	}
}

func TestEnrichContextWithMessageMetadata(t *testing.T) {
	_ = enrichContextWithMessageMetadata
}

func TestFormatHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers []*struct{ Key, Value []byte }
		want    string
	}{
		{
			name:    "no_headers",
			headers: nil,
			want:    "none",
		},
		{
			name:    "empty_headers",
			headers: []*struct{ Key, Value []byte }{},
			want:    "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatHeaders(nil)
			if got != tt.want {
				t.Errorf("formatHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
