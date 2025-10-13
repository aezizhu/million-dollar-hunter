package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AWSClient struct {
	base baseClient
	sm   *secretsmanager.Client
}

type AWSConfig struct {
	Config
	Region string
}

func NewAWS(ctx context.Context, cfg AWSConfig) (*AWSClient, error) {
	if cfg.Region == "" {
		return nil, errors.New("aws region required")
	}
	awsc, err := awsCfg.LoadDefaultConfig(ctx, awsCfg.WithRegion(cfg.Region))
	if err != nil {
		return nil, err
	}
	return &AWSClient{
		base: newBase(cfg.Config),
		sm:   secretsmanager.NewFromConfig(awsc),
	}, nil
}

func (c *AWSClient) fetch(ctx context.Context, name string, stage string) (string, error) {
	out, err := c.sm.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     &name,
		VersionStage: &stage,
	})
	if err != nil {
		return "", err
	}
	if out.SecretString == nil {
		return "", ErrNotFound
	}
	return *out.SecretString, nil
}

func (c *AWSClient) Get(ctx context.Context, name string) (string, error) {
	if v, ok := c.base.getCached(name); ok {
		return v, nil
	}
	s, err := c.fetch(ctx, name, "AWSCURRENT")
	if err != nil {
		return "", err
	}
	c.base.setCached(name, s)
	return s, nil
}

func (c *AWSClient) GetJSON(ctx context.Context, name string, v any) error {
	s, err := c.Get(ctx, name)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), v)
}

func (c *AWSClient) StartBackgroundRefresh(ctx context.Context) {
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

func (c *AWSClient) Health(ctx context.Context) error {
	_, err := c.sm.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
		MaxResults: awsInt32(1),
	})
	return err
}

func awsInt32(v int32) *int32 { return &v }
