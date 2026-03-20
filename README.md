# rate-limiter-redis

Strongly Consistent rate limiting library for Go, backed by Redis.

Unlike P2P or eventually consistent limiters, `rate-limiter-redis` uses Atomic Lua Scripting to ensure that your limits are strictly enforced across any number of distributed service instances.

## Features
- Atomic Operations: Zero race conditions via server-side Lua execution.
- Sliding Window Logic: Accurate counting using Redis Hashes.
- Fail-Open Protection: Integrated Circuit Breaker ensures your API stays up even if Redis goes down.
- Thread-Safe: Designed for high-concurrency Go environments.

## Getting Started

### Installation

```
go get github.com/Hiroki111/rate-limiter-redis
```

### Library Usage (Go Middleware)

```
import (
    "github.com/Hiroki111/rate-limiter-redis/pkg/limiter"
    "github.com/redis/go-redis/v9"
)

func main() {
    // 1. Initialize Redis Client
    rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    
    // 2. Create the Limiter (Limit: 100 requests, Window: 60 seconds)
    engine := limiter.NewRedisLimiter(rdb, 100, 60)
    
    // 3. Use in your HTTP Handler
    http.HandleFunc("/api/resource", func(w http.ResponseWriter, r *http.Request) {
        userID := r.URL.Query().Get("user")
        
        allowed, err := engine.Allow(r.Context(), userID)
        if err != nil || !allowed {
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }
        
        // Process request...
    })
}
```

See `cmd/server/main.go` to see a working example.

## Local Development & Testing

### Running with Docker

```
docker run --name rate-limit-redis -p 6379:6379 -d redis
go run cmd/server/main.go --limit=5

# Open a new terminal and run this to test the limiter
for i in {1..6}; do curl "http://localhost:8080/api/resource?user=alice"; echo ""; done
```

### Running with Docker Compose
```
# Start Redis and the Limiter Service
docker-compose up --build

# Test the limit (set to 5 in docker-compose.yml)
for i in {1..6}; do curl "http://localhost:8080/api/resource?user=alice"; echo ""; done
```

### Running Unit Tests

```
go test -v ./pkg/limiter/...
```