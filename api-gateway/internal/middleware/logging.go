package middleware

import (
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const (
	maxScanLength = 1024 * 1024
)

var (
	piiScrubbingEnabled = os.Getenv("ENABLE_PII_SCRUBBING") == "true"
)

var (
	emailRe    = regexp.MustCompile(`[a-zA-Z0-9._%+\-]{1,64}@[a-zA-Z0-9.\-]{1,255}\.[a-zA-Z]{2,}`)
	uuidRe     = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`)
	walletRe   = regexp.MustCompile(`0x[0-9a-fA-F]{40}`)
	jwtRe      = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{4,}\.[a-zA-Z0-9_-]{4,}\.[a-zA-Z0-9_-]{3,}`)
	passKeysRe = regexp.MustCompile(`(?i)(password|pass|pwd|secret|token|authorization|api_key|api-key)=[^&\n\r]{1,256}`)
)

func maskIP(ip string) string {
	if ip == "" {
		return ip
	}
	
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	
	if parsed.To4() != nil {
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			return parts[0] + "." + parts[1] + "." + parts[2] + ".xxx"
		}
	} else {
		idx := strings.LastIndex(ip, ":")
		if idx != -1 {
			return ip[:idx+1] + "xxxx"
		}
	}
	
	return ip
}

func scrubString(s string) string {
	if s == "" || len(s) > maxScanLength {
		return s
	}

	out := s
	out = passKeysRe.ReplaceAllStringFunc(out, func(match string) string {
		eqIdx := strings.Index(match, "=")
		if eqIdx > 0 {
			return match[:eqIdx+1] + "[redacted]"
		}
		return match
	})
	out = jwtRe.ReplaceAllString(out, "[redacted_jwt]")
	out = emailRe.ReplaceAllString(out, "user@***")
	out = walletRe.ReplaceAllString(out, "[redacted_wallet]")
	out = uuidRe.ReplaceAllString(out, "[redacted_id]")

	var result strings.Builder
	result.Grow(len(out))
	i := 0
	for i < len(out) {
		if (i < len(out) && out[i] >= '0' && out[i] <= '9') || (i+1 < len(out) && out[i:i+2] == "::") {
			j := i
			for j < len(out) && (out[j] >= '0' && out[j] <= '9' || out[j] == '.' || out[j] == ':' || (out[j] >= 'a' && out[j] <= 'f') || (out[j] >= 'A' && out[j] <= 'F')) {
				j++
			}
			if j > i {
				candidate := out[i:j]
				if parsed := net.ParseIP(candidate); parsed != nil {
					result.WriteString(maskIP(candidate))
					i = j
					continue
				}
			}
		}
		result.WriteByte(out[i])
		i++
	}

	return result.String()
}

func safeUserAgent(ua string) string {
	return scrubString(ua)
}

func safePath(path, raw string) string {
	p := path
	if raw != "" {
		p = p + "?" + raw
	}
	return scrubString(p)
}

func Logging(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		traceID := ""
		if v, ok := c.Get("request_id"); ok {
			if tid, ok2 := v.(string); ok2 {
				traceID = tid
			}
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := maskIP(c.ClientIP())
		userAgent := safeUserAgent(c.Request.UserAgent())
		scrubbedPath := safePath(path, raw)

		ev := logger.Info()
		if status >= 400 {
			ev = logger.Error()
		}

		ev.
			Str("trace_id", traceID).
			Str("method", method).
			Str("path", scrubbedPath).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Str("user_agent", userAgent).
			Msg("HTTP request")

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error().
					Str("trace_id", traceID).
					Str("path", scrubbedPath).
					Str("error", scrubString(e.Error())).
					Msg("Request error")
			}
		}
	}
}
