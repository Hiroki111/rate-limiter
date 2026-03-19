package limiter

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"
)

type RedisRateLimiter struct {
	client *redis.Client
	script *redis.Script
	cb     *gobreaker.CircuitBreaker[int]
	limit  int
	window int
}

func NewRedisLimiter(client *redis.Client, limit int, window int) *RedisRateLimiter {
	var st gobreaker.Settings
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 3 && failureRatio >= 0.5
	}

	cb := gobreaker.NewCircuitBreaker[int](st)
	return &RedisRateLimiter{
		client: client,
		script: redis.NewScript(luaScriptSource),
		cb:     cb,
		limit:  limit,
		window: window,
	}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, userID string) (bool, error) {
	result, err := r.cb.Execute(func() (int, error) {
		now := time.Now().Unix()
		return r.script.Run(ctx, r.client, []string{"limiter:" + userID}, now, r.window, r.limit).Int()
	})

	if err != nil {
		log.Printf("Rate limiter degraded (error: %v). Failing open for user %s", err, userID)
		return true, err
	}

	return result == 1, nil
}
