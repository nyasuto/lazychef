package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"lazychef/internal/models"
	"lazychef/internal/services"
)

// MockRecipeGeneratorService is a mock implementation for testing
type MockRecipeGeneratorService struct {
	mock.Mock
}

func (m *MockRecipeGeneratorService) GenerateRecipe(ctx context.Context, req services.RecipeGenerationRequest) (*services.RecipeGenerationResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*services.RecipeGenerationResult), args.Error(1)
}

func (m *MockRecipeGeneratorService) GenerateBatchRecipes(ctx context.Context, req services.BatchGenerationRequest) (*services.BatchGenerationResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*services.BatchGenerationResult), args.Error(1)
}

func (m *MockRecipeGeneratorService) GetHealth() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockRecipeGeneratorService) ClearCache() {
	m.Called()
}

func TestRecipeHandler_SearchRecipes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock generator service
	mockGenerator := &services.RecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test basic search
	req, _ := http.NewRequest("GET", "/api/recipes/search?tag=簡単&ingredient=豚肉", nil)
	c.Request = req

	// Parse query parameters for the context
	c.Request.URL.RawQuery = "tag=簡単&ingredient=豚肉"
	params, _ := url.ParseQuery(c.Request.URL.RawQuery)
	c.Request.Form = params

	handler.SearchRecipes(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check response structure
	response := w.Body.String()
	assert.Contains(t, response, "success")
	assert.Contains(t, response, "data")
}

func TestRecipeHandler_SearchRecipes_InvalidParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &services.RecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with invalid parameters (limit should be positive)
	req, _ := http.NewRequest("GET", "/api/recipes/search?limit=-1", nil)
	c.Request = req

	// Parse query parameters
	c.Request.URL.RawQuery = "limit=-1"
	params, _ := url.ParseQuery(c.Request.URL.RawQuery)
	c.Request.Form = params

	handler.SearchRecipes(c)

	// Handler should handle this gracefully and use default limit
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecipeHandler_SearchRecipes_EmptyQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &services.RecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/api/recipes/search", nil)
	c.Request = req

	handler.SearchRecipes(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should return mock results
	response := w.Body.String()
	assert.Contains(t, response, "豚キャベツ炒め")
}

func TestRecipeHandler_GenerateRecipe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &MockRecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	// Mock successful response
	expectedResult := &services.RecipeGenerationResult{
		Recipe: &models.RecipeData{
			Title:       "テスト料理",
			CookingTime: 10,
			Ingredients: []models.Ingredient{{Name: "材料1", Amount: "100g"}},
			Steps:       []string{"手順1", "手順2"},
			LazinessScore: 9.0,
		},
		Metadata: map[string]interface{}{"test": true},
	}

	mockGenerator.On("GenerateRecipe", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("services.RecipeGenerationRequest")).Return(expectedResult, nil)

	// Create test request
	reqData := services.RecipeGenerationRequest{
		Ingredients: []string{"豚肉", "キャベツ"},
		Season:      "all",
		Servings:    2,
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", "/api/recipes/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GenerateRecipe(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockGenerator.AssertExpectations(t)
}

func TestRecipeHandler_GenerateRecipe_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &MockRecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	req, _ := http.NewRequest("POST", "/api/recipes/generate", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GenerateRecipe(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request format")
}

func TestRecipeHandler_GenerateBatchRecipes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &MockRecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	expectedResult := &services.BatchGenerationResult{
		Recipes: []*models.RecipeData{
			{Title: "料理1", CookingTime: 10},
			{Title: "料理2", CookingTime: 15},
		},
		Metadata: map[string]interface{}{"count": 2},
	}

	mockGenerator.On("GenerateBatchRecipes", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("services.BatchGenerationRequest")).Return(expectedResult, nil)

	reqData := services.BatchGenerationRequest{
		RecipeGenerationRequest: services.RecipeGenerationRequest{
			Ingredients: []string{"豚肉"},
		},
		Count: 2,
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", "/api/recipes/generate-batch", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GenerateBatchRecipes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockGenerator.AssertExpectations(t)
}

func TestRecipeHandler_GetGeneratorHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &MockRecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	expectedHealth := map[string]interface{}{
		"status": "healthy",
		"cache_hits": 10,
	}

	mockGenerator.On("GetHealth").Return(expectedHealth)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetGeneratorHealth(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	mockGenerator.AssertExpectations(t)
}

func TestRecipeHandler_ClearCache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &MockRecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	mockGenerator.On("ClearCache").Return()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.ClearCache(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Cache cleared successfully")
	mockGenerator.AssertExpectations(t)
}

func TestRecipeHandler_GetCacheStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &MockRecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetCacheStats(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Cache stats endpoint")
}

func TestRecipeHandler_TestRecipeGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGenerator := &MockRecipeGeneratorService{}
	handler := NewRecipeHandler(mockGenerator)

	expectedResult := &services.RecipeGenerationResult{
		Recipe: &models.RecipeData{
			Title:       "テスト料理",
			CookingTime: 10,
		},
		Metadata: map[string]interface{}{"test": true},
	}

	mockGenerator.On("GenerateRecipe", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("services.RecipeGenerationRequest")).Return(expectedResult, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/api/recipes/test?ingredients=豚肉&season=spring&cooking_time=15", nil)
	c.Request = req

	handler.TestRecipeGeneration(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test recipe generation successful")
	mockGenerator.AssertExpectations(t)
}
