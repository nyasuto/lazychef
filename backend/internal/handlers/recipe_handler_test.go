package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"lazychef/internal/services"
)

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
