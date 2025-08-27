package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"lazychef/internal/models"
)

func TestNewRecipeCache(t *testing.T) {
	cache := NewRecipeCache(100, time.Hour)

	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestRecipeCache_SetAndGet(t *testing.T) {
	cache := NewRecipeCache(100, time.Hour)

	result := &GenerationResult{
		Recipe: &models.RecipeData{
			Title:         "Test Recipe",
			CookingTime:   15,
			Ingredients:   []models.Ingredient{{Name: "ingredient1", Amount: "1 cup"}},
			Steps:         []string{"step1"},
			Season:        "all",
			LazinessScore: 8.0,
			ServingSize:   2,
		},
		Metadata: GenerationMetadata{
			RequestID:   "test-123",
			Model:       "gpt-3.5-turbo",
			TokensUsed:  100,
			GeneratedAt: time.Now(),
		},
	}

	// Set value
	cache.Set("test-key", result)
	assert.Equal(t, 1, cache.Size())

	// Get value
	retrieved := cache.Get("test-key")
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test Recipe", retrieved.Recipe.Title)
	assert.Equal(t, "test-123", retrieved.Metadata.RequestID)
}

func TestRecipeCache_GetNonExistent(t *testing.T) {
	cache := NewRecipeCache(100, time.Hour)

	result := cache.Get("nonexistent-key")
	assert.Nil(t, result)
}

func TestRecipeCache_Expiration(t *testing.T) {
	// Create cache with very short expiration
	cache := NewRecipeCache(100, 10*time.Millisecond)

	result := &GenerationResult{
		Recipe: &models.RecipeData{
			Title: "Test Recipe",
		},
	}

	cache.Set("test-key", result)
	assert.Equal(t, 1, cache.Size())

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	retrieved := cache.Get("test-key")
	assert.Nil(t, retrieved) // Should be expired
}

func TestRecipeCache_MaxSize(t *testing.T) {
	// Create small cache
	cache := NewRecipeCache(2, time.Hour)

	result1 := &GenerationResult{Recipe: &models.RecipeData{Title: "Recipe 1"}}
	result2 := &GenerationResult{Recipe: &models.RecipeData{Title: "Recipe 2"}}
	result3 := &GenerationResult{Recipe: &models.RecipeData{Title: "Recipe 3"}}

	cache.Set("key1", result1)
	cache.Set("key2", result2)
	assert.Equal(t, 2, cache.Size())

	// Adding third item should evict oldest
	cache.Set("key3", result3)
	assert.Equal(t, 2, cache.Size()) // Still 2 due to max size

	// key1 should be evicted
	assert.Nil(t, cache.Get("key1"))
	assert.NotNil(t, cache.Get("key2"))
	assert.NotNil(t, cache.Get("key3"))
}

func TestRecipeCache_Clear(t *testing.T) {
	cache := NewRecipeCache(100, time.Hour)

	result := &GenerationResult{
		Recipe: &models.RecipeData{Title: "Test Recipe"},
	}

	cache.Set("test-key", result)
	assert.Equal(t, 1, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())
	assert.Nil(t, cache.Get("test-key"))
}

func TestRecipeCache_UpdateExisting(t *testing.T) {
	cache := NewRecipeCache(100, time.Hour)

	result1 := &GenerationResult{
		Recipe: &models.RecipeData{Title: "Original Recipe"},
	}
	result2 := &GenerationResult{
		Recipe: &models.RecipeData{Title: "Updated Recipe"},
	}

	cache.Set("test-key", result1)
	cache.Set("test-key", result2) // Update same key

	assert.Equal(t, 1, cache.Size()) // Size should stay the same

	retrieved := cache.Get("test-key")
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Updated Recipe", retrieved.Recipe.Title)
}
