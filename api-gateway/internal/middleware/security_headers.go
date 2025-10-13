package middleware

import (
	"strconv"
	"strings"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/pkg/headers"
	"github.com/gin-gonic/gin"
)

func SecurityHeaders(cfg config.Config) gin.HandlerFunc {
	hstsValue := ""
	if cfg.HSTSEnabled {
		maxAge := cfg.HSTSMaxAge
		if maxAge <= 0 {
			maxAge = 15552000
		}
		parts := []string{"max-age=" + strconv.Itoa(maxAge)}
		if cfg.HSTSIncludeSubdomains {
			parts = append(parts, "includeSubDomains")
		}
		if cfg.HSTSPreload {
			parts = append(parts, "preload")
		}
		hstsValue = strings.Join(parts, "; ")
	}
	csp := cfg.CSPPolicy
	if strings.ContainsAny(csp, "\r\n") {
		csp = ""
	}

	return func(c *gin.Context) {
		if hstsValue != "" && strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			c.Header(headers.StrictTransportSecurity, hstsValue)
		}
		c.Header(headers.XContentTypeOptions, "nosniff")
		c.Header(headers.ReferrerPolicy, "strict-origin-when-cross-origin")
		c.Header(headers.XFrameOptions, "DENY")
		if csp != "" {
			c.Header(headers.ContentSecurityPolicy, csp)
		}
		if v := cfg.CORP; v != "" {
			c.Header("Cross-Origin-Resource-Policy", v)
		}
		if v := cfg.COOP; v != "" {
			c.Header("Cross-Origin-Opener-Policy", v)
		}
		if v := cfg.PermissionsPolicy; v != "" {
			c.Header("Permissions-Policy", v)
		}
		c.Next()
	}
}
