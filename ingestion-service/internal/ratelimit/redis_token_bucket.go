package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type TokenBucket struct {
	rdb       *redis.Client
	provider  string
	bucket    string
	rate      int
	capacity  int
}

func New(rdb *redis.Client, provider, bucket string, rate, capacity int) *TokenBucket {
	return &TokenBucket{rdb: rdb, provider: provider, bucket: bucket, rate: rate, capacity: capacity}
}

func key(provider, bucket string) string {
	return fmt.Sprintf("ratelimit:%s:%s", provider, bucket)
}

func (t *TokenBucket) Allow(ctx context.Context) (bool, time.Duration, error) {
	k := key(t.provider, t.bucket)
	now := time.Now().Unix()
	pipe := t.rdb.TxPipeline()
	tokens := pipe.HGet(ctx, k, "tokens")
	last := pipe.HGet(ctx, k, "last_refill_ts")
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, 0, err
	}
	var curTokens int
	var lastRefill int64
	if v, err := strconv.Atoi(tokens.Val()); err == nil {
		curTokens = v
	}
	if v, err := strconv.ParseInt(last.Val(), 10, 64); err == nil {
		lastRefill = v
	}
	if lastRefill == 0 {
		lastRefill = now
		curTokens = t.capacity
	}
	elapsed := now - lastRefill
	refill := int(elapsed) * t.rate
	if refill > 0 {
		curTokens = min(t.capacity, curTokens+refill)
		lastRefill = now
	}
	if curTokens <= 0 {
		_ = t.rdb.HSet(ctx, k, "tokens", curTokens, "last_refill_ts", lastRefill).Err()
		return false, time.Second, nil
	}
	curTokens--
	_ = t.rdb.HSet(ctx, k, "tokens", curTokens, "last_refill_ts", lastRefill).Err()
	return true, 0, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
