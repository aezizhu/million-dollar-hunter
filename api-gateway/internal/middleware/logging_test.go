package middleware

import "testing"

func TestScrubEmail(t *testing.T) {
	in := "contact me at alice@example.com"
	out := scrubString(in)
	if out != "contact me at user@***" {
		t.Fatalf("expected email redacted, got %q", out)
	}
}

func TestMaskIPv4(t *testing.T) {
	in := "ip=192.168.1.42"
	out := scrubString(in)
	if out != "ip=192.168.1.xxx" {
		t.Fatalf("expected ipv4 masked, got %q", out)
	}
}

func TestMaskIPv6(t *testing.T) {
	in := "addr=2001:db8:85a3::8a2e:370:7334"
	out := scrubString(in)
	if out == in {
		t.Fatalf("expected ipv6 to be scrubbed")
	}
	if out == "addr=2001:db8:85a3::8a2e:370:xxxx" {
	} else if out == "addr=2001:db8:85a3::8a2e:xxxx" {
	} else {
		if out[len(out)-4:] != "xxxx" {
			t.Fatalf("expected ipv6 last hextet masked, got %q", out)
		}
	}
}

func TestScrubJWT(t *testing.T) {
	in := "jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0In0.sgn"
	out := scrubString(in)
	if out == in || out == "" {
		t.Fatalf("expected jwt redacted, got %q", out)
	}
}

func TestNoOverRedaction(t *testing.T) {
	in := "path=/api/healthz?version=1"
	out := scrubString(in)
	if out != in {
		t.Fatalf("expected no change for benign input, got %q", out)
	}
}
