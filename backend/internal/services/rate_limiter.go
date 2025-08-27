package services

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a rate limiter for API calls
type RateLimiter struct {
	tokens  chan struct{}
	ticker  *time.Ticker
	rate    int
	mu      sync.Mutex
	stopped bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, requestsPerMinute),
		ticker: time.NewTicker(time.Minute / time.Duration(requestsPerMinute)),
		rate:   requestsPerMinute,
	}

	// Fill initial tokens
	for i := 0; i < requestsPerMinute; i++ {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Channel is full, stop filling
			return rl
		}
	}

	// Start the token replenishment goroutine
	go rl.refillTokens()

	return rl
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryWait tries to get a token without waiting
func (rl *RateLimiter) TryWait() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// refillTokens adds tokens to the bucket at the specified rate
func (rl *RateLimiter) refillTokens() {
	for range rl.ticker.C {
		rl.mu.Lock()
		if rl.stopped {
			rl.mu.Unlock()
			return
		}
		rl.mu.Unlock()

		// Try to add a token
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full, skip
		}
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.stopped {
		rl.stopped = true
		rl.ticker.Stop()
		close(rl.tokens)
	}
}

// GetAvailableTokens returns the number of available tokens
func (rl *RateLimiter) GetAvailableTokens() int {
	return len(rl.tokens)
}

// GetRate returns the configured rate
func (rl *RateLimiter) GetRate() int {
	return rl.rate
}
