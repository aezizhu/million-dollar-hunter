package secrets

import (
	"context"
	"encoding/json"
	"os"
)

type envClient struct {
	baseClient
}

func NewEnv(cfg Config) Client {
	return &envClient{baseClient: newBase(cfg)}
}

func (e *envClient) Get(ctx context.Context, name string) (string, error) {
	if v, ok := e.getCached(name); ok {
		return v, nil
	}
	v := os.Getenv(toEnvKey(name))
	if v == "" {
		return "", ErrNotFound
	}
	e.setCached(name, v)
	return v, nil
}

func (e *envClient) GetJSON(ctx context.Context, name string, v any) error {
	s, err := e.Get(ctx, name)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), v)
}

func (e *envClient) StartBackgroundRefresh(ctx context.Context) {}

func (e *envClient) Health(ctx context.Context) error { return nil }

func toEnvKey(name string) string {
	k := name
	for i := 0; i < len(k); i++ {
		if k[i] == '/' {
			k = k[:i] + "_" + k[i+1:]
		}
	}
	return stringsToUpper(k)
}

func stringsToUpper(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] = b[i] - 32
		}
	}
	return string(b)
}
