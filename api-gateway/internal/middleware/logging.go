package middleware

import (
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var (
	emailRe    = regexp.MustCompile(`([a-zA-Z0-9._%+\-]+)@([a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})`)
	uuidRe     = regexp.MustCompile(`\b[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[1-5][0-9a-fA-F]{3}\-[89abAB][0-9a-fA-F]{3}\-[0-9a-fA-F]{12}\b`)
	walletRe   = regexp.MustCompile(`\b0x[0-9a-fA-F]{40}\b`)
	jwtRe      = regexp.MustCompile(`\beyJ[a-zA-Z0-9_\-]+?\.[a-zA-Z0-9_\-]+?\.[a-zA-Z0-9_\-]+?\b`)
	passKeysRe = regexp.MustCompile(`(?i)(password|pass|pwd|secret|token|authorization|api_key|api-key)=([^&\s]+)`)
	ipv4Re     = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3})\.\d{1,3}\b`)
)

func maskIP(ip string) string {
	if ip == "" {
		return ip
	}
	return ipv4Re.ReplaceAllString(ip, `$1.xxx`)
}

func scrubString(s string) string {
	if s == "" {
		return s
	}
	out := s
	out = passKeysRe.ReplaceAllString(out, `${1}=[redacted]`)
	out = jwtRe.ReplaceAllString(out, "[redacted_jwt]")
	out = emailRe.ReplaceAllString(out, "user@***")
	out = walletRe.ReplaceAllString(out, "[redacted_wallet]")
	out = uuidRe.ReplaceAllString(out, "[redacted_id]")
	out = ipv4Re.ReplaceAllString(out, `$1.xxx`)
	return out
}

func safeUserAgent(ua string) string {
	return scrubString(ua)
}

func safePath(path, raw string) string {
	p := path
	if raw != "" {
		p = p + "?" + raw
	}
	p = scrubString(p)
	if strings.Contains(strings.ToLower(p), "authorization=") {
		p = passKeysRe.ReplaceAllString(p, `${1}=[redacted]`)
	}
	return p
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
