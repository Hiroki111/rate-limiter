package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"rate-limiter-redis/internal/limiter"

	"github.com/redis/go-redis/v9"
)

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	redisAddr := flag.String("redis-addr", "localhost:6379", "Redis address")
	limit := flag.Int("limit", 100, "Global requests per minute")
	flag.Parse()

	fmt.Printf("Starting at :%s\n", *port)
	redisClient := redis.NewClient(&redis.Options{Addr: *redisAddr})
	engine := limiter.NewRedisLimiter(redisClient, *limit, 60)

	http.HandleFunc("/api/resource", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user")
		if userID == "" {
			userID = "default-user"
		}

		allowed, err := engine.Allow(r.Context(), userID)
		if err != nil {
			// Fail-open logic could be added here
			http.Error(w, "Internal Limiter Error", 500)
			return
		}

		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, "Rate limit exceeded. Try again later.")
			return
		}

		fmt.Fprint(w, "Access Granted")
	})

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
