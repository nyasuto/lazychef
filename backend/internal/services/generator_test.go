package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"lazychef/internal/config"
	"lazychef/internal/models"
)

func TestNewRecipeGeneratorService(t *testing.T) {
	config := &config.OpenAIConfig{
		APIKey:      "test-key",
		Model:       "gpt-3.5-turbo",
		MaxTokens:   1000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
	}

	service := NewRecipeGeneratorService(config)

	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
	assert.NotNil(t, service.config)
	assert.NotNil(t, service.rateLimiter)
	assert.NotNil(t, service.cache)
}

func TestRecipeGenerationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     RecipeGenerationRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: RecipeGenerationRequest{
				Ingredients:    []string{"豚肉", "キャベツ"},
				Season:         "spring",
				MaxCookingTime: 15,
				Servings:       2,
				Tags:           []string{"簡単"},
			},
			wantErr: false,
		},
		{
			name: "empty ingredients",
			req: RecipeGenerationRequest{
				Ingredients: []string{},
				Season:      "spring",
			},
			wantErr: true,
		},
		{
			name: "invalid season",
			req: RecipeGenerationRequest{
				Ingredients: []string{"豚肉"},
				Season:      "invalid",
			},
			wantErr: true,
		},
		{
			name: "zero cooking time",
			req: RecipeGenerationRequest{
				Ingredients:    []string{"豚肉"},
				Season:         "spring",
				MaxCookingTime: 0,
			},
			wantErr: true,
		},
		{
			name: "zero servings",
			req: RecipeGenerationRequest{
				Ingredients: []string{"豚肉"},
				Season:      "spring",
				Servings:    0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBatchGenerationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     BatchGenerationRequest
		wantErr bool
	}{
		{
			name: "valid batch request",
			req: BatchGenerationRequest{
				RecipeGenerationRequest: RecipeGenerationRequest{
					Ingredients:    []string{"豚肉"},
					Season:         "spring",
					MaxCookingTime: 15,
					Servings:       2,
				},
				Count: 3,
			},
			wantErr: false,
		},
		{
			name: "zero count",
			req: BatchGenerationRequest{
				RecipeGenerationRequest: RecipeGenerationRequest{
					Ingredients: []string{"豚肉"},
					Season:      "spring",
				},
				Count: 0,
			},
			wantErr: true,
		},
		{
			name: "count too high",
			req: BatchGenerationRequest{
				RecipeGenerationRequest: RecipeGenerationRequest{
					Ingredients: []string{"豚肉"},
					Season:      "spring",
				},
				Count: 11,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecipeGeneratorService_GetHealth(t *testing.T) {
	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	service := NewRecipeGeneratorService(config)
	health := service.GetHealth()

	assert.NotNil(t, health)
	assert.Contains(t, health, "status")
	assert.Contains(t, health, "cache_size")
	assert.Contains(t, health, "rate_limiter_active")
	assert.Equal(t, "healthy", health["status"])
}

func TestRecipeGeneratorService_ClearCache(t *testing.T) {
	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	service := NewRecipeGeneratorService(config)
	
	// Add something to cache first
	testRecipe := &models.RecipeData{Title: "Test"}
	service.cache.Set("test-key", testRecipe, time.Hour)
	
	// Verify cache has item
	_, found := service.cache.Get("test-key")
	assert.True(t, found)
	
	// Clear cache
	service.ClearCache()
	
	// Verify cache is empty
	_, found = service.cache.Get("test-key")
	assert.False(t, found)
}

func TestGenerationMetadata(t *testing.T) {
	metadata := GenerationMetadata{
		RequestID:      "test-123",
		Model:          "gpt-3.5-turbo",
		TokensUsed:     100,
		GeneratedAt:    time.Now(),
		ProcessingTime: 2 * time.Second,
		CacheHit:       true,
		RetryCount:     0,
	}

	assert.Equal(t, "test-123", metadata.RequestID)
	assert.Equal(t, "gpt-3.5-turbo", metadata.Model)
	assert.Equal(t, 100, metadata.TokensUsed)
	assert.True(t, metadata.CacheHit)
	assert.Equal(t, 0, metadata.RetryCount)
}

func TestGenerationResult_Structure(t *testing.T) {
	recipe := &models.RecipeData{
		Title:       "Test Recipe",
		CookingTime: 10,
	}

	result := GenerationResult{
		Recipe: recipe,
		Error:  "",
		Metadata: GenerationMetadata{
			RequestID: "test",
		},
	}

	assert.Equal(t, "Test Recipe", result.Recipe.Title)
	assert.Empty(t, result.Error)
	assert.Equal(t, "test", result.Metadata.RequestID)
}

// Integration test for cache key generation
func TestGenerateCacheKey(t *testing.T) {
	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	service := NewRecipeGeneratorService(config)
	
	req := RecipeGenerationRequest{
		Ingredients: []string{"豚肉", "キャベツ"},
		Season:      "spring",
		Servings:    2,
	}

	key1 := service.generateCacheKey(req)
	key2 := service.generateCacheKey(req)

	// Same request should generate same key
	assert.Equal(t, key1, key2)
	assert.NotEmpty(t, key1)

	// Different request should generate different key
	req2 := req
	req2.Ingredients = []string{"鶏肉"}
	key3 := service.generateCacheKey(req2)
	assert.NotEqual(t, key1, key3)
}

// Test context cancellation behavior
func TestRecipeGeneratorService_ContextCancellation(t *testing.T) {
	config := &config.OpenAIConfig{
		APIKey:      "test-key",
		Model:       "gpt-3.5-turbo",
		MaxTokens:   1000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
	}

	service := NewRecipeGeneratorService(config)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := RecipeGenerationRequest{
		Ingredients:    []string{"豚肉"},
		Season:         "spring",
		MaxCookingTime: 15,
		Servings:       2,
	}

	// This should fail quickly due to cancelled context
	result, err := service.GenerateRecipe(ctx, req)
	
	// Should either get an error or a result with error
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotEmpty(t, result.Error)
	}
}