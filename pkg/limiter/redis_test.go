package limiter

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisLimiter_Allow(t *testing.T) {
	// 1. Start miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// 2. Setup the client and limiter
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	limit := 5
	window := 60
	l := NewRedisLimiter(client, limit, window)
	mockTime := time.Now().Unix()
	l.now = func() int64 {
		return mockTime
	}
	ctx := context.Background()
	userID := "test-user"

	// 3. Test: First 5 requests should be allowed
	for i := 1; i <= limit; i++ {
		allowed, err := l.Allow(ctx, userID)
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Errorf("request %d: expected allowed, got blocked", i)
		}
	}

	// 4. Test: The 6th request should be blocked
	allowed, err := l.Allow(ctx, userID)
	if err != nil {
		t.Fatalf("6th request: unexpected error: %v", err)
	}
	if allowed {
		t.Error("6th request: expected blocked, got allowed")
	}

	// 5. Test: Fast-forward time to check window reset
	mockTime += int64(window + 1)
	mr.FastForward(time.Duration(window+1) * time.Second)

	// 6. Test: Should be allowed again after window expires
	allowed, err = l.Allow(ctx, userID)
	if !allowed || err != nil {
		t.Error("expected allowed after window reset, got blocked")
	}
}

func TestRedisLimiter_Concurrency(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	limit := 10
	l := NewRedisLimiter(client, limit, 60)
	ctx := context.Background()
	userID := "concurrent-user"

	var wg sync.WaitGroup
	var allowedCount int32
	totalRequests := 50

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, err := l.Allow(ctx, userID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if allowed {
				atomic.AddInt32(&allowedCount, 1)
			}
		}()
	}

	wg.Wait()

	if int(allowedCount) != limit {
		t.Errorf("Expected exactly %d allowed requests, but got %d", limit, allowedCount)
	}
}
