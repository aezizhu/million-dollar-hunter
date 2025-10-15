package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
)

type awsClient struct {
	baseClient
	sm     *secretsmanager.Client
	prefix string
}

type AWSConfig struct {
	Config Config
	Region string
}

func NewAWS(ctx context.Context, cfg AWSConfig) (Client, error) {
	region := cfg.Region
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	awscfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	sm := secretsmanager.NewFromConfig(awscfg)
	prefix := os.Getenv("SECRETS_PREFIX")
	if prefix == "" {
		env := os.Getenv("ENV")
		if env == "" {
			env = "dev"
		}
		prefix = "mdh/" + env
	}
	return &awsClient{
		baseClient: newBase(cfg.Config),
		sm:         sm,
		prefix:     prefix,
	}, nil
}

func (a *awsClient) fullName(name string) string {
	if name == "" {
		return a.prefix
	}
	if a.prefix == "" {
		return name
	}
	return a.prefix + "/" + name
}

func (a *awsClient) fetch(ctx context.Context, name string) (string, error) {
	out, err := a.sm.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &name,
	})
	if err != nil {
		var ae smithy.APIError
		if ok := errors.As(err, &ae); ok && ae.ErrorCode() == "ResourceNotFoundException" {
			return "", ErrNotFound
		}
		return "", err
	}
	if out.SecretString == nil {
		return "", ErrNotFound
	}
	return *out.SecretString, nil
}

func (a *awsClient) Get(ctx context.Context, name string) (string, error) {
	if v, ok := a.getCached(name); ok {
		return v, nil
	}
	full := a.fullName(name)
	s, err := a.fetch(ctx, full)
	if err != nil {
		return "", err
	}
	a.setCached(name, s)
	return s, nil
}

func (a *awsClient) GetJSON(ctx context.Context, name string, v any) error {
	s, err := a.Get(ctx, name)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), v)
}

func (a *awsClient) StartBackgroundRefresh(ctx context.Context) {
	t := time.NewTicker(a.cfg.RefreshInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				a.refreshOnce(ctx)
			}
		}
	}()
}

func (a *awsClient) refreshOnce(ctx context.Context) {
	a.mu.RLock()
	keys := make([]string, 0, len(a.cache))
	for k := range a.cache {
		keys = append(keys, k)
	}
	a.mu.RUnlock()
	for _, k := range keys {
		_, _ = a.Get(ctx, k)
	}
}

func (a *awsClient) Health(ctx context.Context) error {
	_, err := a.Get(ctx, "")
	if err == nil || err == ErrNotFound {
		return nil
	}
	return err
}
