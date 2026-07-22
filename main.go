package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (b *TokenBucket) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()

	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

var (
	buckets   = make(map[string]*TokenBucket)
	bucketsMu sync.Mutex // protects buckets map from concurrent access
)

func getBucket(clientID string) *TokenBucket {
	bucketsMu.Lock()
	defer bucketsMu.Unlock()

	b, exists := buckets[clientID]
	if !exists {
		b = NewTokenBucket(10, 2)
		buckets[clientID] = b
	}
	return b
}

func startCleanup(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // short for testing; use 30s+ in real use
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Cleanup goroutine stopped")
			return
		case <-ticker.C:
			bucketsMu.Lock()
			for id, b := range buckets {
				if time.Since(b.lastRefill) > 15*time.Second { // idle threshold; short for testing
					delete(buckets, id)
					fmt.Printf("Evicted idle client: %s\n", id)
				}
			}
			bucketsMu.Unlock()
		}
	}
}

func limitHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client")
	if clientID == "" {
		clientID = "anonymous"
	}

	bucket := getBucket(clientID)

	if bucket.Allow() {
		fmt.Fprintf(w, "allowed for %s\n", clientID)
	} else {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, "rejected for %s\n", clientID)
	}
}

func main() {
	http.HandleFunc("/request", limitHandler)
	fmt.Println("Server running on :8080")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startCleanup(ctx)
	http.ListenAndServe(":8080", nil)
}
