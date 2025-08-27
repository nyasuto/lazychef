package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"lazychef/internal/config"
	"lazychef/internal/models"
)

// RecipeGeneratorService handles AI-powered recipe generation
type RecipeGeneratorService struct {
	client      *openai.Client
	config      *config.OpenAIConfig
	rateLimiter *RateLimiter
	cache       *RecipeCache
}

// GenerationResult holds the result of recipe generation
type GenerationResult struct {
	Recipe   *models.RecipeData  `json:"recipe,omitempty"`
	Recipes  []models.RecipeData `json:"recipes,omitempty"`
	Error    string              `json:"error,omitempty"`
	Metadata GenerationMetadata  `json:"metadata"`
}

// GenerationMetadata holds metadata about the generation process
type GenerationMetadata struct {
	RequestID      string        `json:"request_id"`
	Model          string        `json:"model"`
	TokensUsed     int           `json:"tokens_used"`
	GeneratedAt    time.Time     `json:"generated_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	CacheHit       bool          `json:"cache_hit"`
	RetryCount     int           `json:"retry_count"`
}

// BatchGenerationRequest represents a request for multiple recipes
type BatchGenerationRequest struct {
	RecipeGenerationRequest
	Count int `json:"count" binding:"required,min=1,max=10"`
}

// NewRecipeGeneratorService creates a new recipe generator service
func NewRecipeGeneratorService(config *config.OpenAIConfig) (*RecipeGeneratorService, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	client := openai.NewClient(config.APIKey)
	rateLimiter := NewRateLimiter(config.RequestsPerMinute)
	cache := NewRecipeCache(1000, 24*time.Hour) // Cache for 24 hours

	return &RecipeGeneratorService{
		client:      client,
		config:      config,
		rateLimiter: rateLimiter,
		cache:       cache,
	}, nil
}

// GenerateRecipe generates a single recipe based on the request
func (s *RecipeGeneratorService) GenerateRecipe(ctx context.Context, req RecipeGenerationRequest) (*GenerationResult, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	// Check cache first
	cacheKey := s.generateCacheKey(req)
	if cachedResult := s.cache.Get(cacheKey); cachedResult != nil {
		cachedResult.Metadata.CacheHit = true
		cachedResult.Metadata.RequestID = requestID
		cachedResult.Metadata.ProcessingTime = time.Since(startTime)
		return cachedResult, nil
	}

	// Rate limiting
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting error: %w", err)
	}

	// Generate prompt
	promptTemplate := GetRecipeGenerationPrompt(req)

	result := &GenerationResult{
		Metadata: GenerationMetadata{
			RequestID:   requestID,
			Model:       s.config.Model,
			GeneratedAt: time.Now(),
			CacheHit:    false,
		},
	}

	// Call OpenAI API with retries
	recipe, tokensUsed, retryCount, err := s.callOpenAIWithRetry(ctx, promptTemplate)
	if err != nil {
		result.Error = err.Error()
		result.Metadata.ProcessingTime = time.Since(startTime)
		result.Metadata.RetryCount = retryCount
		return result, err
	}

	result.Recipe = recipe
	result.Metadata.TokensUsed = tokensUsed
	result.Metadata.ProcessingTime = time.Since(startTime)
	result.Metadata.RetryCount = retryCount

	// Validate and enhance the generated recipe
	if err := s.validateAndEnhanceRecipe(result.Recipe); err != nil {
		return result, fmt.Errorf("recipe validation failed: %w", err)
	}

	// Cache the result
	s.cache.Set(cacheKey, result)

	log.Printf("Generated recipe '%s' in %v (tokens: %d, retries: %d)",
		result.Recipe.Title, result.Metadata.ProcessingTime, tokensUsed, retryCount)

	return result, nil
}

// GenerateBatchRecipes generates multiple recipes based on the request
func (s *RecipeGeneratorService) GenerateBatchRecipes(ctx context.Context, req BatchGenerationRequest) (*GenerationResult, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	// Check cache
	cacheKey := s.generateBatchCacheKey(req)
	if cachedResult := s.cache.Get(cacheKey); cachedResult != nil {
		cachedResult.Metadata.CacheHit = true
		cachedResult.Metadata.RequestID = requestID
		cachedResult.Metadata.ProcessingTime = time.Since(startTime)
		return cachedResult, nil
	}

	// Rate limiting
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting error: %w", err)
	}

	// Generate prompt for batch generation
	promptTemplate := GetBatchRecipeGenerationPrompt(req.RecipeGenerationRequest, req.Count)

	result := &GenerationResult{
		Metadata: GenerationMetadata{
			RequestID:   requestID,
			Model:       s.config.Model,
			GeneratedAt: time.Now(),
			CacheHit:    false,
		},
	}

	// Call OpenAI API
	recipes, tokensUsed, retryCount, err := s.callOpenAIBatchWithRetry(ctx, promptTemplate)
	if err != nil {
		result.Error = err.Error()
		result.Metadata.ProcessingTime = time.Since(startTime)
		result.Metadata.RetryCount = retryCount
		return result, err
	}

	result.Recipes = recipes
	result.Metadata.TokensUsed = tokensUsed
	result.Metadata.ProcessingTime = time.Since(startTime)
	result.Metadata.RetryCount = retryCount

	// Validate each recipe
	for i := range result.Recipes {
		if err := s.validateAndEnhanceRecipe(&result.Recipes[i]); err != nil {
			log.Printf("Warning: Recipe %d validation failed: %v", i, err)
		}
	}

	// Cache the result
	s.cache.Set(cacheKey, result)

	log.Printf("Generated %d recipes in %v (tokens: %d, retries: %d)",
		len(recipes), result.Metadata.ProcessingTime, tokensUsed, retryCount)

	return result, nil
}

// callOpenAIWithRetry calls OpenAI API with retry logic for single recipe
func (s *RecipeGeneratorService) callOpenAIWithRetry(ctx context.Context, prompt PromptTemplate) (*models.RecipeData, int, int, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(s.config.RetryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, 0, attempt, ctx.Err()
			}
		}

		recipe, tokensUsed, err := s.callOpenAIForRecipe(ctx, prompt)
		if err == nil {
			return recipe, tokensUsed, attempt, nil
		}

		lastErr = err
		log.Printf("Recipe generation attempt %d failed: %v", attempt+1, err)

		// Don't retry on certain errors
		if isNonRetryableError(err) {
			break
		}
	}

	return nil, 0, s.config.MaxRetries, fmt.Errorf("failed after %d attempts: %w", s.config.MaxRetries+1, lastErr)
}

// callOpenAIBatchWithRetry calls OpenAI API with retry logic for batch recipes
func (s *RecipeGeneratorService) callOpenAIBatchWithRetry(ctx context.Context, prompt PromptTemplate) ([]models.RecipeData, int, int, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(s.config.RetryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, 0, attempt, ctx.Err()
			}
		}

		recipes, tokensUsed, err := s.callOpenAIForBatchRecipes(ctx, prompt)
		if err == nil {
			return recipes, tokensUsed, attempt, nil
		}

		lastErr = err
		log.Printf("Batch recipe generation attempt %d failed: %v", attempt+1, err)

		if isNonRetryableError(err) {
			break
		}
	}

	return nil, 0, s.config.MaxRetries, fmt.Errorf("failed after %d attempts: %w", s.config.MaxRetries+1, lastErr)
}

// callOpenAIForRecipe makes the actual API call for single recipe
func (s *RecipeGeneratorService) callOpenAIForRecipe(ctx context.Context, prompt PromptTemplate) (*models.RecipeData, int, error) {
	req := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		Temperature: s.config.Temperature,
		MaxTokens:   s.config.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt.SystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt.UserPrompt,
			},
		},
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	resp, err := s.client.CreateChatCompletion(timeoutCtx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, resp.Usage.TotalTokens, errors.New("no choices returned from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Parse JSON response
	var recipe models.RecipeData
	if err := json.Unmarshal([]byte(content), &recipe); err != nil {
		return nil, resp.Usage.TotalTokens, fmt.Errorf("failed to parse recipe JSON: %w", err)
	}

	return &recipe, resp.Usage.TotalTokens, nil
}

// callOpenAIForBatchRecipes makes the actual API call for batch recipes
func (s *RecipeGeneratorService) callOpenAIForBatchRecipes(ctx context.Context, prompt PromptTemplate) ([]models.RecipeData, int, error) {
	req := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		Temperature: s.config.Temperature,
		MaxTokens:   s.config.MaxTokens * 3, // More tokens for multiple recipes
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt.SystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt.UserPrompt,
			},
		},
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	resp, err := s.client.CreateChatCompletion(timeoutCtx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, resp.Usage.TotalTokens, errors.New("no choices returned from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Parse JSON response
	var batchResponse struct {
		Recipes []models.RecipeData `json:"recipes"`
	}

	if err := json.Unmarshal([]byte(content), &batchResponse); err != nil {
		return nil, resp.Usage.TotalTokens, fmt.Errorf("failed to parse batch recipe JSON: %w", err)
	}

	return batchResponse.Recipes, resp.Usage.TotalTokens, nil
}

// validateAndEnhanceRecipe validates and enhances the generated recipe
func (s *RecipeGeneratorService) validateAndEnhanceRecipe(recipe *models.RecipeData) error {
	if recipe == nil {
		return errors.New("recipe is nil")
	}

	// Validate required fields
	if err := recipe.Validate(); err != nil {
		return err
	}

	// Enhance recipe with calculated laziness score if not present or seems wrong
	calculatedScore := recipe.CalculateLazinessScore()
	if recipe.LazinessScore < 1.0 || recipe.LazinessScore > 10.0 {
		recipe.LazinessScore = calculatedScore
	} else {
		// Use average of AI score and calculated score for better accuracy
		recipe.LazinessScore = (recipe.LazinessScore + calculatedScore) / 2.0
	}

	// Set default serving size if not specified
	if recipe.ServingSize <= 0 {
		recipe.ServingSize = 1
	}

	// Set default difficulty if not specified
	if recipe.Difficulty == "" {
		if recipe.LazinessScore >= 8.0 {
			recipe.Difficulty = "easy"
		} else if recipe.LazinessScore >= 6.0 {
			recipe.Difficulty = "medium"
		} else {
			recipe.Difficulty = "hard"
		}
	}

	return nil
}

// Helper functions

func (s *RecipeGeneratorService) generateCacheKey(req RecipeGenerationRequest) string {
	return fmt.Sprintf("recipe:%s:%s:%d:%d",
		strings.Join(req.Ingredients, ","),
		req.Season,
		req.MaxCookingTime,
		req.Servings,
	)
}

func (s *RecipeGeneratorService) generateBatchCacheKey(req BatchGenerationRequest) string {
	return fmt.Sprintf("batch:%s:%s:%d:%d:%d",
		strings.Join(req.Ingredients, ","),
		req.Season,
		req.MaxCookingTime,
		req.Servings,
		req.Count,
	)
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func isNonRetryableError(err error) bool {
	// Check for specific error types that shouldn't be retried
	errStr := err.Error()
	return strings.Contains(errStr, "invalid_api_key") ||
		strings.Contains(errStr, "insufficient_quota") ||
		strings.Contains(errStr, "invalid_request_error")
}

// GetHealth returns the health status of the service
func (s *RecipeGeneratorService) GetHealth() map[string]interface{} {
	return map[string]interface{}{
		"status":          "healthy",
		"model":           s.config.Model,
		"cache_size":      s.cache.Size(),
		"rate_limit_rpm":  s.config.RequestsPerMinute,
		"max_retries":     s.config.MaxRetries,
		"request_timeout": s.config.RequestTimeout.String(),
	}
}

// ClearCache clears the recipe cache
func (s *RecipeGeneratorService) ClearCache() {
	s.cache.Clear()
}
