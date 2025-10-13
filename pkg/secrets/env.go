package secrets

import (
	"context"
	"encoding/json"
	"os"
	"time"
)

type EnvClient struct {
	base baseClient
}

func NewEnv(cfg Config) *EnvClient {
	return &EnvClient{base: newBase(cfg)}
}

func (c *EnvClient) Get(ctx context.Context, name string) (string, error) {
	if v, ok := c.base.getCached(name); ok {
		return v, nil
	}
	v := os.Getenv(name)
	if v == "" {
		v = os.Getenv(ToEnvKey(name))
	}
	if v == "" {
		return "", ErrNotFound
	}
	c.base.setCached(name, v)
	return v, nil
}

func (c *EnvClient) GetJSON(ctx context.Context, name string, v any) error {
	s, err := c.Get(ctx, name)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), v)
}

func (c *EnvClient) StartBackgroundRefresh(ctx context.Context) {
	t := time.NewTicker(c.base.cfg.RefreshInterval)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
}

func (c *EnvClient) Health(ctx context.Context) error {
	return nil
}
