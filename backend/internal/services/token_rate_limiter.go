package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

// TokenRateLimiter provides advanced token-level rate limiting and cost control
type TokenRateLimiter struct {
	requestLimiter *RateLimiter
	tokenBucket    *TokenBucket
	metrics        *TokenMetrics
	metricsMu      sync.RWMutex
	backoffConfig  *BackoffConfig
	costTracker    *CostTracker
	costTrackerMu  sync.RWMutex
	mu             sync.RWMutex
}

// TokenBucket implements token bucket algorithm for token-level rate limiting
type TokenBucket struct {
	capacity   int     // Maximum tokens in bucket
	tokens     float64 // Current tokens available
	refillRate float64 // Tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// TokenMetrics tracks detailed token usage statistics
type TokenMetrics struct {
	TotalRequests        int64   `json:"total_requests"`
	TotalTokensUsed      int64   `json:"total_tokens_used"`
	PromptTokensUsed     int64   `json:"prompt_tokens_used"`
	CompletionTokensUsed int64   `json:"completion_tokens_used"`
	EstimatedCostUSD     float64 `json:"estimated_cost_usd"`
	RequestsBlocked      int64   `json:"requests_blocked"`
	TokensRejected       int64   `json:"tokens_rejected"`
	AverageLatency       float64 `json:"average_latency_ms"`
	RateLimitHits        int64   `json:"rate_limit_hits"`
}

// BackoffConfig configures exponential backoff for rate limit errors
type BackoffConfig struct {
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	Multiplier    float64       `json:"multiplier"`
	MaxRetries    int           `json:"max_retries"`
	JitterEnabled bool          `json:"jitter_enabled"`
}

// CostTracker monitors and controls costs
type CostTracker struct {
	DailyBudgetUSD   float64   `json:"daily_budget_usd"`
	MonthlyBudgetUSD float64   `json:"monthly_budget_usd"`
	DailySpentUSD    float64   `json:"daily_spent_usd"`
	MonthlySpentUSD  float64   `json:"monthly_spent_usd"`
	LastResetDaily   time.Time `json:"last_reset_daily"`
	LastResetMonthly time.Time `json:"last_reset_monthly"`
	AlertThreshold   float64   `json:"alert_threshold"` // 0.8 = 80% of budget
}

// TokenUsageEstimate represents estimated resource usage for a request
type TokenUsageEstimate struct {
	EstimatedPromptTokens     int     `json:"estimated_prompt_tokens"`
	EstimatedCompletionTokens int     `json:"estimated_completion_tokens"`
	EstimatedTotalTokens      int     `json:"estimated_total_tokens"`
	EstimatedCostUSD          float64 `json:"estimated_cost_usd"`
	Model                     string  `json:"model"`
}

// RateLimitResult represents the result of a rate limiting check
type RateLimitResult struct {
	Allowed         bool          `json:"allowed"`
	Reason          string        `json:"reason,omitempty"`
	RetryAfter      time.Duration `json:"retry_after,omitempty"`
	TokensAvailable int           `json:"tokens_available"`
	CostRemaining   float64       `json:"cost_remaining_usd"`
}

// NewTokenRateLimiter creates a new advanced token rate limiter
func NewTokenRateLimiter(requestsPerMinute int, tokensPerSecond int, dailyBudgetUSD, monthlyBudgetUSD float64) *TokenRateLimiter {
	return &TokenRateLimiter{
		requestLimiter: NewRateLimiter(requestsPerMinute),
		tokenBucket: &TokenBucket{
			capacity:   tokensPerSecond * 60, // 1 minute worth of tokens
			tokens:     float64(tokensPerSecond * 60),
			refillRate: float64(tokensPerSecond),
			lastRefill: time.Now(),
		},
		metrics: &TokenMetrics{},
		backoffConfig: &BackoffConfig{
			InitialDelay:  time.Second,
			MaxDelay:      5 * time.Minute,
			Multiplier:    2.0,
			MaxRetries:    5,
			JitterEnabled: true,
		},
		costTracker: &CostTracker{
			DailyBudgetUSD:   dailyBudgetUSD,
			MonthlyBudgetUSD: monthlyBudgetUSD,
			LastResetDaily:   time.Now().Truncate(24 * time.Hour),
			LastResetMonthly: time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC),
			AlertThreshold:   0.8,
		},
	}
}

// CheckRateLimit performs comprehensive rate limiting checks before making a request
func (t *TokenRateLimiter) CheckRateLimit(ctx context.Context, estimate TokenUsageEstimate) *RateLimitResult {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Update cost tracker (daily/monthly reset check)
	t.updateCostTracker()

	// Check budget constraints
	t.costTrackerMu.RLock()
	dailyBudget := t.costTracker.DailyBudgetUSD
	dailySpent := t.costTracker.DailySpentUSD
	monthlyBudget := t.costTracker.MonthlyBudgetUSD
	monthlySpent := t.costTracker.MonthlySpentUSD
	t.costTrackerMu.RUnlock()

	if dailyBudget > 0 {
		if dailySpent+estimate.EstimatedCostUSD > dailyBudget {
			return &RateLimitResult{
				Allowed:       false,
				Reason:        "daily budget exceeded",
				CostRemaining: math.Max(0, dailyBudget-dailySpent),
			}
		}
	}

	if monthlyBudget > 0 {
		if monthlySpent+estimate.EstimatedCostUSD > monthlyBudget {
			return &RateLimitResult{
				Allowed:       false,
				Reason:        "monthly budget exceeded",
				CostRemaining: math.Max(0, monthlyBudget-monthlySpent),
			}
		}
	}

	// Check token bucket
	available := t.tokenBucket.getAvailableTokens()
	if available < estimate.EstimatedTotalTokens {
		retryAfter := t.calculateRetryAfter(estimate.EstimatedTotalTokens - available)
		return &RateLimitResult{
			Allowed:         false,
			Reason:          "token rate limit exceeded",
			RetryAfter:      retryAfter,
			TokensAvailable: available,
		}
	}

	// Check request rate limiter
	if !t.requestLimiter.TryWait() {
		return &RateLimitResult{
			Allowed:    false,
			Reason:     "request rate limit exceeded",
			RetryAfter: time.Minute,
		}
	}

	return &RateLimitResult{
		Allowed:         true,
		TokensAvailable: available,
		CostRemaining:   math.Max(0, dailyBudget-dailySpent),
	}
}

// WaitForCapacity waits until sufficient capacity is available (with context cancellation)
func (t *TokenRateLimiter) WaitForCapacity(ctx context.Context, estimate TokenUsageEstimate) error {
	for {
		result := t.CheckRateLimit(ctx, estimate)
		if result.Allowed {
			return nil
		}

		// If budget exceeded, don't wait
		if result.Reason == "daily budget exceeded" || result.Reason == "monthly budget exceeded" {
			return fmt.Errorf("budget exceeded: %s", result.Reason)
		}

		// Wait before retrying
		waitTime := result.RetryAfter
		if waitTime == 0 {
			waitTime = time.Second // Default wait time
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue to next iteration
		}
	}
}

// ReserveTokens attempts to reserve tokens for a request
func (t *TokenRateLimiter) ReserveTokens(estimate TokenUsageEstimate) bool {
	t.tokenBucket.mu.Lock()
	defer t.tokenBucket.mu.Unlock()

	t.tokenBucket.refill()

	if t.tokenBucket.tokens >= float64(estimate.EstimatedTotalTokens) {
		t.tokenBucket.tokens -= float64(estimate.EstimatedTotalTokens)
		return true
	}

	return false
}

// RecordUsage records actual token usage and cost after a successful request
func (t *TokenRateLimiter) RecordUsage(usage openai.Usage, actualCostUSD float64) {
	t.metricsMu.Lock()
	t.metrics.TotalRequests++
	t.metrics.TotalTokensUsed += int64(usage.TotalTokens)
	t.metrics.PromptTokensUsed += int64(usage.PromptTokens)
	t.metrics.CompletionTokensUsed += int64(usage.CompletionTokens)
	t.metrics.EstimatedCostUSD += actualCostUSD
	t.metricsMu.Unlock()

	// Update cost tracker
	t.costTrackerMu.Lock()
	t.costTracker.DailySpentUSD += actualCostUSD
	t.costTracker.MonthlySpentUSD += actualCostUSD
	t.costTrackerMu.Unlock()
}

// RecordRateLimitHit records when a 429 error occurs
func (t *TokenRateLimiter) RecordRateLimitHit() {
	t.metricsMu.Lock()
	t.metrics.RateLimitHits++
	t.metricsMu.Unlock()
}

// EstimateTokenUsage estimates token usage for a request
func (t *TokenRateLimiter) EstimateTokenUsage(messages []openai.ChatCompletionMessage, model string, maxTokens int) TokenUsageEstimate {
	// Rough estimation based on character count
	// This is a simplified estimation - in production you'd use tiktoken or similar

	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}

	// Rough estimation: ~4 characters per token
	estimatedPromptTokens := totalChars / 4

	// Use maxTokens as completion estimate, or default based on model
	estimatedCompletionTokens := maxTokens
	if estimatedCompletionTokens == 0 {
		if strings.Contains(model, "gpt-4") {
			estimatedCompletionTokens = 500
		} else {
			estimatedCompletionTokens = 300
		}
	}

	totalTokens := estimatedPromptTokens + estimatedCompletionTokens
	estimatedCost := t.estimateCost(model, estimatedPromptTokens, estimatedCompletionTokens)

	return TokenUsageEstimate{
		EstimatedPromptTokens:     estimatedPromptTokens,
		EstimatedCompletionTokens: estimatedCompletionTokens,
		EstimatedTotalTokens:      totalTokens,
		EstimatedCostUSD:          estimatedCost,
		Model:                     model,
	}
}

// GetMetrics returns current token usage metrics
func (t *TokenRateLimiter) GetMetrics() TokenMetrics {
	t.metricsMu.RLock()
	defer t.metricsMu.RUnlock()
	// Create a copy to avoid returning a structure with a mutex
	return TokenMetrics{
		TotalRequests:        t.metrics.TotalRequests,
		TotalTokensUsed:      t.metrics.TotalTokensUsed,
		PromptTokensUsed:     t.metrics.PromptTokensUsed,
		CompletionTokensUsed: t.metrics.CompletionTokensUsed,
		EstimatedCostUSD:     t.metrics.EstimatedCostUSD,
		RequestsBlocked:      t.metrics.RequestsBlocked,
		TokensRejected:       t.metrics.TokensRejected,
		AverageLatency:       t.metrics.AverageLatency,
		RateLimitHits:        t.metrics.RateLimitHits,
	}
}

// GetCostStatus returns current cost tracking status
func (t *TokenRateLimiter) GetCostStatus() CostTracker {
	t.costTrackerMu.RLock()
	defer t.costTrackerMu.RUnlock()
	// Create a copy to avoid returning a structure with a mutex
	return CostTracker{
		DailyBudgetUSD:   t.costTracker.DailyBudgetUSD,
		MonthlyBudgetUSD: t.costTracker.MonthlyBudgetUSD,
		DailySpentUSD:    t.costTracker.DailySpentUSD,
		MonthlySpentUSD:  t.costTracker.MonthlySpentUSD,
		LastResetDaily:   t.costTracker.LastResetDaily,
		LastResetMonthly: t.costTracker.LastResetMonthly,
		AlertThreshold:   t.costTracker.AlertThreshold,
	}
}

// SetBudgets updates daily and monthly budgets
func (t *TokenRateLimiter) SetBudgets(dailyUSD, monthlyUSD float64) {
	t.costTrackerMu.Lock()
	t.costTracker.DailyBudgetUSD = dailyUSD
	t.costTracker.MonthlyBudgetUSD = monthlyUSD
	t.costTrackerMu.Unlock()
}

// ResetDailyUsage manually resets daily usage (for testing)
func (t *TokenRateLimiter) ResetDailyUsage() {
	t.costTrackerMu.Lock()
	t.costTracker.DailySpentUSD = 0
	t.costTracker.LastResetDaily = time.Now().Truncate(24 * time.Hour)
	t.costTrackerMu.Unlock()
}

// Helper methods

func (tb *TokenBucket) getAvailableTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return int(tb.tokens)
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	if elapsed > 0 {
		tokensToAdd := elapsed * tb.refillRate
		tb.tokens = math.Min(tb.tokens+tokensToAdd, float64(tb.capacity))
		tb.lastRefill = now
	}
}

func (t *TokenRateLimiter) calculateRetryAfter(tokensNeeded int) time.Duration {
	// Calculate how long to wait for enough tokens
	tokensPerSecond := t.tokenBucket.refillRate
	secondsNeeded := float64(tokensNeeded) / tokensPerSecond
	return time.Duration(secondsNeeded * float64(time.Second))
}

func (t *TokenRateLimiter) updateCostTracker() {
	now := time.Now()

	t.costTrackerMu.Lock()
	defer t.costTrackerMu.Unlock()

	// Check if we need to reset daily usage
	if now.Truncate(24 * time.Hour).After(t.costTracker.LastResetDaily) {
		t.costTracker.DailySpentUSD = 0
		t.costTracker.LastResetDaily = now.Truncate(24 * time.Hour)
	}

	// Check if we need to reset monthly usage
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if monthStart.After(t.costTracker.LastResetMonthly) {
		t.costTracker.MonthlySpentUSD = 0
		t.costTracker.LastResetMonthly = monthStart
	}
}

func (t *TokenRateLimiter) estimateCost(model string, promptTokens, completionTokens int) float64 {
	// Cost estimation based on OpenAI pricing (as of 2024)
	// These would need to be updated with current pricing

	var promptCostPer1k, completionCostPer1k float64

	switch {
	case strings.Contains(model, "gpt-4-turbo"):
		promptCostPer1k = 0.01     // $0.01 per 1K prompt tokens
		completionCostPer1k = 0.03 // $0.03 per 1K completion tokens
	case strings.Contains(model, "gpt-4"):
		promptCostPer1k = 0.03     // $0.03 per 1K prompt tokens
		completionCostPer1k = 0.06 // $0.06 per 1K completion tokens
	case strings.Contains(model, "gpt-3.5-turbo"):
		promptCostPer1k = 0.001     // $0.001 per 1K prompt tokens
		completionCostPer1k = 0.002 // $0.002 per 1K completion tokens
	case strings.HasPrefix(model, "gpt-5"):
		// GPT-5 pricing - estimated
		promptCostPer1k = 0.05     // $0.05 per 1K prompt tokens
		completionCostPer1k = 0.15 // $0.15 per 1K completion tokens
	default:
		// Default to GPT-3.5-turbo pricing
		promptCostPer1k = 0.001
		completionCostPer1k = 0.002
	}

	promptCost := (float64(promptTokens) / 1000) * promptCostPer1k
	completionCost := (float64(completionTokens) / 1000) * completionCostPer1k

	return promptCost + completionCost
}

// HandleRateLimitError implements exponential backoff for 429 errors
func (t *TokenRateLimiter) HandleRateLimitError(ctx context.Context, attempt int) error {
	t.RecordRateLimitHit()

	if attempt >= t.backoffConfig.MaxRetries {
		return fmt.Errorf("max retries exceeded for rate limit")
	}

	delay := time.Duration(float64(t.backoffConfig.InitialDelay) * math.Pow(t.backoffConfig.Multiplier, float64(attempt)))
	if delay > t.backoffConfig.MaxDelay {
		delay = t.backoffConfig.MaxDelay
	}

	// Add jitter to prevent thundering herd
	if t.backoffConfig.JitterEnabled {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1) // 10% jitter
		delay += jitter
	}

	log.Printf("Rate limit hit, waiting %v before retry (attempt %d/%d)", delay, attempt+1, t.backoffConfig.MaxRetries)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// Stop gracefully shuts down the rate limiter
func (t *TokenRateLimiter) Stop() {
	if t.requestLimiter != nil {
		t.requestLimiter.Stop()
	}
}
