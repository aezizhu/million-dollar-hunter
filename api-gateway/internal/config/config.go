package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type Config struct {
	Port                    string
	RedisURL                string
	AuthMode                string
	AdminUser               string
	AdminPass               string
	AuthServiceURL          string
	PortfolioServiceURL     string
	MarketDataServiceURL    string
	AuthValidateMode        string
	AuthGRPCAddr            string
	AuthGRPCTimeoutMs       int
	AuthGRPCFallbackToLocal bool
	JWTSecret               string
	JWTAudience             string
	FrontendURL             string
	OTLPEndpoint            string
	PrometheusNamespace     string
	RateDefaultRPS          int
	RateDefaultBurst        int
	OpenAPIPath             string
	RouteLimitsJSON         string
	StrictOpenAPIValidation bool

	RateLimitAllowlist    string
	RateLimitBypassHeader string
	IPRateLimitRPS        int
	IPRateLimitBurst      int
	UserRateLimitRPS      int
	UserRateLimitBurst    int
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getenvInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func getenvBool(k string, d bool) bool {
	if v := os.Getenv(k); v != "" {
		if v == "1" || v == "true" || v == "TRUE" {
			return true
		}
		if v == "0" || v == "false" || v == "FALSE" {
			return false
		}
	}
	return d
}

func Load() Config {
	return Config{
		Port:                    getenv("PORT", "8080"),
		RedisURL:                getenv("REDIS_URL", "localhost:6379"),
		AuthMode:                getenv("AUTH_MODE", "mvp-gate"),
		AdminUser:               getenv("ADMIN_USER", ""),
		AdminPass:               getenv("ADMIN_PASS", ""),
		AuthServiceURL:          os.Getenv("AUTH_SERVICE_URL"),
		PortfolioServiceURL:     os.Getenv("PORTFOLIO_SERVICE_URL"),
		MarketDataServiceURL:    os.Getenv("MARKET_DATA_SERVICE_URL"),
		AuthValidateMode:        getenv("AUTH_VALIDATE_MODE", "local"),
		AuthGRPCAddr:            os.Getenv("AUTH_GRPC_ADDR"),
		AuthGRPCTimeoutMs:       getenvInt("AUTH_GRPC_TIMEOUT_MS", 2000),
		AuthGRPCFallbackToLocal: getenvBool("AUTH_GRPC_FALLBACK_TO_LOCAL", false),
		JWTSecret:               os.Getenv("JWT_SECRET"),
		JWTAudience:             os.Getenv("JWT_AUDIENCE"),
		FrontendURL:             getenv("FRONTEND_URL", "*"),
		OTLPEndpoint:            os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		PrometheusNamespace:     getenv("PROMETHEUS_NAMESPACE", "api_gateway"),
		RateDefaultRPS:          getenvInt("RATE_DEFAULT_RPS", 10),
		RateDefaultBurst:        getenvInt("RATE_DEFAULT_BURST", 20),
		OpenAPIPath:             getenv("OPENAPI_PATH", "../docs/openapi.yaml"),
		RouteLimitsJSON:         os.Getenv("ROUTE_LIMITS"),
		StrictOpenAPIValidation: getenvBool("STRICT_OPENAPI_VALIDATION", false),

		RateLimitAllowlist:    os.Getenv("RATE_LIMIT_ALLOWLIST"),
		RateLimitBypassHeader: os.Getenv("RATE_LIMIT_BYPASS_HEADER"),
		IPRateLimitRPS:        getenvInt("IP_RATE_LIMIT_RPS", 0),
		IPRateLimitBurst:      getenvInt("IP_RATE_LIMIT_BURST", 0),
		UserRateLimitRPS:      getenvInt("USER_RATE_LIMIT_RPS", 0),
		UserRateLimitBurst:    getenvInt("USER_RATE_LIMIT_BURST", 0),
	}

func (c Config) Validate(logger zerolog.Logger) {
	if c.IPRateLimitRPS < 0 {
		logger.Warn().Str("env", "IP_RATE_LIMIT_RPS").Msg("negative value; treating as 0 (inherit defaults)")
	}
	if c.IPRateLimitBurst < 0 {
		logger.Warn().Str("env", "IP_RATE_LIMIT_BURST").Msg("negative value; treating as 0 (inherit defaults)")
	}
	if c.UserRateLimitRPS < 0 {
		logger.Warn().Str("env", "USER_RATE_LIMIT_RPS").Msg("negative value; treating as 0 (inherit defaults)")
	}
	if c.UserRateLimitBurst < 0 {
		logger.Warn().Str("env", "USER_RATE_LIMIT_BURST").Msg("negative value; treating as 0 (inherit defaults)")
	}
	if c.IPRateLimitBurst > 0 && c.IPRateLimitRPS > c.IPRateLimitBurst {
		logger.Warn().Int("rps", c.IPRateLimitRPS).Int("burst", c.IPRateLimitBurst).Msg("IP_RATE_LIMIT_RPS exceeds IP_RATE_LIMIT_BURST")
	}
	if c.UserRateLimitBurst > 0 && c.UserRateLimitRPS > c.UserRateLimitBurst {
		logger.Warn().Int("rps", c.UserRateLimitRPS).Int("burst", c.UserRateLimitBurst).Msg("USER_RATE_LIMIT_RPS exceeds USER_RATE_LIMIT_BURST")
	}
	logger.Info().
		Int("ip_rps", c.IPRateLimitRPS).Int("ip_burst", c.IPRateLimitBurst).
		Int("user_rps", c.UserRateLimitRPS).Int("user_burst", c.UserRateLimitBurst).
		Msg("effective rate limits")
}
