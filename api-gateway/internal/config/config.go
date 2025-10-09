package config

import (
	"os"
)

type Config struct {
	Port                   string
	RedisURL               string
	AuthMode               string
	AdminUser              string
	AdminPass              string
	OTLPEndpoint           string
	PrometheusNamespace    string
	RateDefaultRPS         int
	RateDefaultBurst       int
	OpenAPIPath            string
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func Load() Config {
	return Config{
		Port:                getenv("PORT", "8080"),
		RedisURL:            getenv("REDIS_URL", "localhost:6379"),
		AuthMode:            getenv("AUTH_MODE", "mvp-gate"),
		AdminUser:           getenv("ADMIN_USER", "aezi"),
		AdminPass:           getenv("ADMIN_PASS", "Aa@123456789"),
		OTLPEndpoint:        os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		PrometheusNamespace: getenv("PROMETHEUS_NAMESPACE", "api_gateway"),
		RateDefaultRPS:      10,
		RateDefaultBurst:    20,
		OpenAPIPath:         getenv("OPENAPI_PATH", "../../docs/openapi.yaml"),
	}
}
