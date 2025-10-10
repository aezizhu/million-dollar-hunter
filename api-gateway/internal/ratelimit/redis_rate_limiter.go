package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb      *redis.Client
	rate     int           // tokens per second
	capacity int           // max burst size
	script   *redis.Script
}

const rateLimitScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])

local tokens = tonumber(redis.call('HGET', key, 'tokens') or capacity)
local last_refill = tonumber(redis.call('HGET', key, 'last_refill_ts') or now)

local elapsed = now - last_refill
local refill = math.floor(elapsed) * rate
if refill > 0 then
    tokens = math.min(capacity, tokens + refill)
    last_refill = now
end

if tokens <= 0 then
    redis.call('HSET', key, 'tokens', tokens, 'last_refill_ts', last_refill)
    redis.call('EXPIRE', key, 3600)
    return {0, math.ceil(1 / rate * 1000)}
end

tokens = tokens - 1
redis.call('HSET', key, 'tokens', tokens, 'last_refill_ts', last_refill)
redis.call('EXPIRE', key, 3600)
return {1, 0}
`

func New(rdb *redis.Client, rate, capacity int) *RateLimiter {
	return &RateLimiter{
		rdb:      rdb,
		rate:     rate,
		capacity: capacity,
		script:   redis.NewScript(rateLimitScript),
	}
}

func (r *RateLimiter) Allow(ctx context.Context, identifier string) (allowed bool, retryAfter time.Duration, remaining int, resetTime int64, err error) {
	key := fmt.Sprintf("ratelimit:gateway:%s", identifier)
	now := time.Now().Unix()
	
	result, err := r.script.Run(ctx, r.rdb, []string{key}, now, r.rate, r.capacity).Result()
	if err != nil {
		return false, 0, 0, 0, fmt.Errorf("rate limit script failed: %w", err)
	}
	
	vals, ok := result.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, 0, 0, fmt.Errorf("unexpected script result format")
	}
	
	allowedInt, ok := vals[0].(int64)
	if !ok {
		return false, 0, 0, 0, fmt.Errorf("unexpected allowed value type")
	}
	
	waitMs, ok := vals[1].(int64)
	if !ok {
		return false, 0, 0, 0, fmt.Errorf("unexpected wait value type")
	}
	
	allowed = allowedInt == 1
	retryAfter = time.Duration(waitMs) * time.Millisecond
	
	if allowed {
		pipe := r.rdb.Pipeline()
		tokensCmd := pipe.HGet(ctx, key, "tokens")
		_, _ = pipe.Exec(ctx)
		
		if tokensStr, err := tokensCmd.Result(); err == nil {
			var tokens int64
			fmt.Sscanf(tokensStr, "%d", &tokens)
			remaining = int(tokens)
		}
	}
	
	resetTime = now + int64(r.capacity/r.rate)
	
	return allowed, retryAfter, remaining, resetTime, nil
}

func (r *RateLimiter) Reset(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("ratelimit:gateway:%s", identifier)
	return r.rdb.Del(ctx, key).Err()
}

func (r *RateLimiter) Capacity() int {
	return r.capacity
}

// Rate returns the refill rate per second
func (r *RateLimiter) Rate() int {
	return r.rate
}
