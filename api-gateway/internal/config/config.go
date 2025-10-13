package config

import (
	"os"
	"strconv"
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
	EnableHSTS              bool
	HSTSEnabled             bool
	HSTSMaxAge              int
	HSTSIncludeSubdomains   bool
	HSTSPreload             bool
	CSPPolicy               string
	CORP                    string
	COOP                    string
	PermissionsPolicy       string
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
	hstsEnabled := getenvBool("HSTS_ENABLED", getenvBool("ENABLE_HSTS", false))
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
		EnableHSTS:              getenvBool("ENABLE_HSTS", false),
		HSTSEnabled:             hstsEnabled,
		HSTSMaxAge:              getenvInt("HSTS_MAX_AGE", 15552000),
		HSTSIncludeSubdomains:   getenvBool("HSTS_INCLUDE_SUBDOMAINS", false),
		HSTSPreload:             getenvBool("HSTS_PRELOAD", false),
		CSPPolicy:               getenv("CSP_POLICY", ""),
		CORP:                    getenv("CROSS_ORIGIN_RESOURCE_POLICY", ""),
		COOP:                    getenv("CROSS_ORIGIN_OPENER_POLICY", ""),
		PermissionsPolicy:       getenv("PERMISSIONS_POLICY", ""),
	}
}
