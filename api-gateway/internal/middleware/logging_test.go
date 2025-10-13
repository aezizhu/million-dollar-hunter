package middleware

import (
	"strings"
	"testing"
)

func TestScrubEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single email", "contact me at alice@example.com", "contact me at user@***"},
		{"multiple emails", "alice@ex.com and bob@ex.com", "user@*** and user@***"},
		{"no email", "no email here", "no email here"},
		{"email in URL", "https://alice@example.com/path", "https://user@***/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestMaskIPv4(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple IPv4", "ip=192.168.1.42", "ip=192.168.1.xxx"},
		{"multiple IPs", "from 10.0.0.1 to 10.0.0.2", "from 10.0.0.xxx to 10.0.0.xxx"},
		{"localhost", "127.0.0.1", "127.0.0.xxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestMaskIPv6(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"full IPv6", "addr=2001:db8:85a3::8a2e:370:7334"},
		{"localhost", "addr=::1"},
		{"compressed", "addr=2001:db8::1"},
		{"link-local", "addr=fe80::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out == tt.input {
				t.Errorf("expected IPv6 to be masked, got unchanged: %q", out)
			}
			if !strings.HasSuffix(out, "xxxx") {
				t.Errorf("expected IPv6 last hextet to end with xxxx, got %q", out)
			}
		})
	}
}

func TestScrubJWT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid JWT", "jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0In0.sgn", "jwt=[redacted_jwt]"},
		{"JWT in header", "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0In0.signature", "Authorization: Bearer [redacted_jwt]"},
		{"no JWT", "no jwt here", "no jwt here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestScrubWallet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Ethereum wallet", "wallet=0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "wallet=0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"}, // 39 chars, not 40
		{"valid wallet", "send to 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", "send to [redacted_wallet]"},                           // 40 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestScrubUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid UUID", "user_id=550e8400-e29b-41d4-a716-446655440000", "user_id=[redacted_id]"},
		{"UUID in path", "/users/550e8400-e29b-41d4-a716-446655440000/profile", "/users/[redacted_id]/profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestScrubSensitiveKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"password", "password=secret123", "password=[redacted]"},
		{"api_key", "api_key=xyz789&other=value", "api_key=[redacted]&other=value"},
		{"token", "token=abc123", "token=[redacted]"},
		{"authorization", "authorization=Bearer xyz", "authorization=[redacted]"},
		{"mixed case", "PASSWORD=test", "PASSWORD=[redacted]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestNoOverRedaction(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"benign path", "path=/api/healthz?version=1"},
		{"normal text", "this is normal text without PII"},
		{"numbers only", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if out != tt.input {
				t.Errorf("expected no change for benign input, got %q", out)
			}
		})
	}
}

func TestReDoSEmailAttack(t *testing.T) {
	malicious := strings.Repeat("a", 50) + "@"

	out := scrubString(malicious)
	if len(out) == 0 {
		t.Error("scrubString returned empty for malicious email input")
	}
}

func TestReDoSLargeInput(t *testing.T) {
	large := strings.Repeat("test@example.com ", 10000)

	out := scrubString(large)
	if len(out) == 0 {
		t.Error("scrubString returned empty for large input")
	}
}

func TestReDoSNestedPatterns(t *testing.T) {
	nested := "aaaaaaaaaa@aaaaaaaaaa@aaaaaaaaaa@"

	out := scrubString(nested)
	if len(out) == 0 {
		t.Error("scrubString returned empty for nested patterns")
	}
}

func TestEmptyString(t *testing.T) {
	out := scrubString("")
	if out != "" {
		t.Errorf("expected empty string, got %q", out)
	}
}

func TestUnicode(t *testing.T) {
	input := "user@例え.jp and 测试@测试.cn"
	out := scrubString(input)
	if len(out) == 0 {
		t.Error("scrubString returned empty for unicode input")
	}
}

func TestExtremelyLongString(t *testing.T) {
	input := strings.Repeat("a", 2*1024*1024) // 2MB
	out := scrubString(input)
	if out != input {
		t.Error("expected unchanged string for input exceeding maxScanLength")
	}
}

func TestMalformedPatterns(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid email", "@invalid"},
		{"no email", "noemail"},
		{"truncated JWT", "eyJhbGci"},
		{"invalid IP", "999.999.999.999"},
		{"incomplete UUID", "550e8400-e29b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := scrubString(tt.input)
			if len(out) == 0 {
				t.Errorf("scrubString returned empty for malformed pattern: %q", tt.input)
			}
		})
	}
}

func TestMultipleOccurrences(t *testing.T) {
	input := "alice@ex.com and bob@ex.com and charlie@ex.com"
	out := scrubString(input)
	expected := "user@*** and user@*** and user@***"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestMaskIPv4Direct(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard IP", "192.168.1.42", "192.168.1.xxx"},
		{"localhost", "127.0.0.1", "127.0.0.xxx"},
		{"empty", "", ""},
		{"invalid", "not-an-ip", "not-an-ip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := maskIP(tt.input)
			if out != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestMaskIPv6Direct(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"full address", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"compressed", "2001:db8::1"},
		{"localhost", "::1"},
		{"link-local", "fe80::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := maskIP(tt.input)
			if out == tt.input {
				t.Errorf("expected IPv6 to be masked, got unchanged: %q", out)
			}
			if !strings.HasSuffix(out, "xxxx") {
				t.Errorf("expected masked IPv6 to end with xxxx, got %q", out)
			}
		})
	}
}

func BenchmarkScrubStringNoMatch(b *testing.B) {
	input := "this is a normal log line without any PII or sensitive data"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubString(input)
	}
}

func BenchmarkScrubStringWithEmail(b *testing.B) {
	input := "user logged in: alice@example.com from 192.168.1.42"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubString(input)
	}
}

func BenchmarkScrubStringWithJWT(b *testing.B) {
	input := "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubString(input)
	}
}

func BenchmarkScrubStringComplex(b *testing.B) {
	input := "user alice@example.com logged in from 192.168.1.42 with token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0In0.sig and wallet 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubString(input)
	}
}

func BenchmarkScrubStringReDoSAttack(b *testing.B) {
	malicious := strings.Repeat("a", 50) + "@"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubString(malicious)
	}
}

func BenchmarkMaskIPv4(b *testing.B) {
	ip := "192.168.1.42"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = maskIP(ip)
	}
}

func BenchmarkMaskIPv6(b *testing.B) {
	ip := "2001:db8:85a3::8a2e:370:7334"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = maskIP(ip)
	}
}
