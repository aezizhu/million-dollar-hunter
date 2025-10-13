package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenBucket struct {
	rdb    *redis.Client
	rate   int
	burst  int
	period time.Duration
	prefix string
}

type AllowState struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetUnix  int64
	RetryAfter time.Duration
}

func NewRedisTokenBucket(rdb *redis.Client, rate, burst int, period time.Duration, prefix string) *RedisTokenBucket {
	if prefix == "" {
		prefix = "ratelimit"
	}
	return &RedisTokenBucket{
		rdb:    rdb,
		rate:   rate,
		burst:  burst,
		period: period,
		prefix: prefix,
	}
}

func (r *RedisTokenBucket) key(k string) string {
	return r.prefix + ":" + k
}

func (r *RedisTokenBucket) Allow(ctx context.Context, key string) (bool, int, int, time.Time, time.Duration) {
	k := r.key(key)
	now := time.Now()

	ttl, err := r.rdb.TTL(ctx, k).Result()
	if err != nil && err != redis.Nil {
		ttl = 0
	}
	if ttl <= 0 {
		if err := r.rdb.Set(ctx, k, r.burst, r.period).Err(); err != nil {
			reset := now.Add(r.period)
			return true, r.rate, r.burst - 1, reset, 0
		}
		ttl = r.period
	}

	remaining, err := r.rdb.Decr(ctx, k).Result()
	if err != nil {
		reset := now.Add(ttl)
		return true, r.rate, r.burst - 1, reset, 0
	}

	if remaining < 0 {
		ttl2, _ := r.rdb.TTL(ctx, k).Result()
		if ttl2 <= 0 {
			ttl2 = time.Second
		}
		reset := now.Add(ttl2)
		return false, r.rate, 0, reset, ttl2
	}

	ttl3, _ := r.rdb.TTL(ctx, k).Result()
	reset := now.Add(ttl3)
	return true, r.rate, int(remaining), reset, 0
}
