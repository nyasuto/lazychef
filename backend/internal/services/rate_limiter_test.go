package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(60)
	assert.NotNil(t, limiter)
}

func TestRateLimiter_Wait_AllowsImmediate(t *testing.T) {
	limiter := NewRateLimiter(60) // 60 requests per minute
	ctx := context.Background()

	// First request should not wait
	start := time.Now()
	err := limiter.Wait(ctx)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 10*time.Millisecond) // Should be immediate
}

func TestRateLimiter_Wait_EnforcesRateLimit(t *testing.T) {
	limiter := NewRateLimiter(2) // Only 2 requests per minute for testing
	ctx := context.Background()

	// First request should be immediate
	err := limiter.Wait(ctx)
	assert.NoError(t, err)

	// Second request should be immediate
	err = limiter.Wait(ctx)
	assert.NoError(t, err)

	// Third request should be delayed
	start := time.Now()
	err = limiter.Wait(ctx)
	duration := time.Since(start)

	assert.NoError(t, err)
	// Should wait approximately 30 seconds (60s/2 requests = 30s interval)
	assert.GreaterOrEqual(t, duration, 25*time.Second)
	assert.Less(t, duration, 35*time.Second)
}

func TestRateLimiter_Wait_ContextCanceled(t *testing.T) {
	limiter := NewRateLimiter(1) // Very restrictive for testing

	// Create a context that will be canceled
	ctx, cancel := context.WithCancel(context.Background())

	// Make first request to consume the token
	err := limiter.Wait(ctx)
	assert.NoError(t, err)

	// Cancel context before second request
	cancel()

	// Second request should fail due to canceled context
	err = limiter.Wait(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestRateLimiter_Wait_ContextTimeout(t *testing.T) {
	limiter := NewRateLimiter(1) // Very restrictive for testing

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Make first request to consume the token
	err := limiter.Wait(context.Background())
	assert.NoError(t, err)

	// Second request should timeout
	start := time.Now()
	err = limiter.Wait(ctx)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
	assert.Less(t, duration, 200*time.Millisecond)
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	limiter := NewRateLimiter(10) // 10 requests per minute
	ctx := context.Background()

	// Test concurrent requests don't interfere with each other
	start := time.Now()

	// Make several concurrent requests
	for i := 0; i < 5; i++ {
		go func() {
			err := limiter.Wait(ctx)
			assert.NoError(t, err)
		}()
	}

	duration := time.Since(start)

	// All requests should complete relatively quickly
	// (they might be serialized but shouldn't take too long)
	assert.Less(t, duration, 2*time.Second)
}
