package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lazychef/internal/models"
	"lazychef/internal/services"
)

// RecipeHandler handles recipe-related HTTP requests
type RecipeHandler struct {
	generatorService         *services.RecipeGeneratorService
	enhancedGeneratorService *services.EnhancedRecipeGeneratorService
}

// NewRecipeHandler creates a new recipe handler
func NewRecipeHandler(generatorService *services.RecipeGeneratorService, enhancedGeneratorService *services.EnhancedRecipeGeneratorService) *RecipeHandler {
	return &RecipeHandler{
		generatorService:         generatorService,
		enhancedGeneratorService: enhancedGeneratorService,
	}
}

// GenerateRecipe handles POST /api/recipes/generate
func (h *RecipeHandler) GenerateRecipe(c *gin.Context) {
	var req services.RecipeGenerationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Generate recipe
	result, err := h.generatorService.GenerateRecipe(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate recipe",
			"details": err.Error(),
		})
		return
	}

	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Recipe generation error",
			"details": result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recipe":   result.Recipe,
		"metadata": result.Metadata,
	})
}

// GenerateRecipeEnhanced generates a recipe using GPT-5 with enhanced validation
func (h *RecipeHandler) GenerateRecipeEnhanced(c *gin.Context) {
	var req services.EnhancedGenerationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set default stage if not specified
	if req.Stage == "" {
		req.Stage = services.StageAuthoring
	}

	// Set default reasoning effort and verbosity if not specified
	if req.ReasoningEffort == "" {
		req.ReasoningEffort = h.enhancedGeneratorService.GetConfig().ReasoningEffort
	}
	if req.Verbosity == "" {
		req.Verbosity = h.enhancedGeneratorService.GetConfig().Verbosity
	}

	// Generate recipe using enhanced service
	result, err := h.enhancedGeneratorService.GenerateRecipeEnhanced(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate enhanced recipe",
			"details": err.Error(),
		})
		return
	}

	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Enhanced recipe generation error",
			"details": result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recipe":             result.Recipe,
		"metadata":           result.Metadata,
		"stage":              result.Stage,
		"model_used":         result.ModelUsed,
		"reasoning_effort":   result.ReasoningEffort,
		"verbosity":          result.Verbosity,
		"structured_outputs": result.StructuredOutputs,
		"safety_check":       result.SafetyCheckResult,
		"quality_check":      result.QualityCheckResult,
	})
}

// ValidateRecipeSafety validates a recipe for food safety compliance
func (h *RecipeHandler) ValidateRecipeSafety(c *gin.Context) {
	var recipe models.RecipeData

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid recipe format",
			"details": err.Error(),
		})
		return
	}

	// Create a temporary enhanced service if needed
	if h.enhancedGeneratorService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Enhanced generator service not available",
		})
		return
	}

	// Validate using food safety validator
	safetyResult, err := h.enhancedGeneratorService.GetFoodSafetyValidator().ValidateRecipe(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Safety validation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"safety_result": safetyResult,
	})
}

// ValidateRecipeQuality validates a recipe for quality compliance
func (h *RecipeHandler) ValidateRecipeQuality(c *gin.Context) {
	var recipe models.RecipeData

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid recipe format",
			"details": err.Error(),
		})
		return
	}

	// Create a temporary enhanced service if needed
	if h.enhancedGeneratorService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Enhanced generator service not available",
		})
		return
	}

	// Validate using quality validator
	qualityResult, err := h.enhancedGeneratorService.GetQualityValidator().ValidateRecipe(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Quality validation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quality_result": qualityResult,
	})
}

// GenerateBatchRecipes handles POST /api/recipes/generate-batch
func (h *RecipeHandler) GenerateBatchRecipes(c *gin.Context) {
	var req services.BatchGenerationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Generate batch recipes
	result, err := h.generatorService.GenerateBatchRecipes(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate batch recipes",
			"details": err.Error(),
		})
		return
	}

	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Batch recipe generation error",
			"details": result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recipes":  result.Recipes,
		"metadata": result.Metadata,
	})
}

// GetGeneratorHealth handles GET /api/recipes/health
func (h *RecipeHandler) GetGeneratorHealth(c *gin.Context) {
	health := h.generatorService.GetHealth()
	c.JSON(http.StatusOK, health)
}

// ClearCache handles POST /api/recipes/clear-cache
func (h *RecipeHandler) ClearCache(c *gin.Context) {
	h.generatorService.ClearCache()
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}

// GetCacheStats handles GET /api/recipes/cache-stats
func (h *RecipeHandler) GetCacheStats(c *gin.Context) {
	// This would need to be implemented in the service
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache stats endpoint - to be implemented",
	})
}

// TestRecipeGeneration handles GET /api/recipes/test - for quick testing
func (h *RecipeHandler) TestRecipeGeneration(c *gin.Context) {
	// Default test request
	req := services.RecipeGenerationRequest{
		Ingredients:    []string{"豚こま肉", "キャベツ"},
		Season:         "all",
		MaxCookingTime: 10,
		Servings:       1,
	}

	// Override with query parameters if provided
	if ingredients := c.Query("ingredients"); ingredients != "" {
		// Simple comma-separated parsing
		// In production, this would be more sophisticated
		req.Ingredients = []string{ingredients}
	}

	if season := c.Query("season"); season != "" {
		req.Season = season
	}

	if cookingTimeStr := c.Query("cooking_time"); cookingTimeStr != "" {
		if cookingTime, err := strconv.Atoi(cookingTimeStr); err == nil {
			req.MaxCookingTime = cookingTime
		}
	}

	// Generate recipe
	result, err := h.generatorService.GenerateRecipe(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Test recipe generation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Test recipe generation successful",
		"request":  req,
		"recipe":   result.Recipe,
		"metadata": result.Metadata,
	})
}

// SearchRecipes handles GET /api/recipes/search
func (h *RecipeHandler) SearchRecipes(c *gin.Context) {
	var criteria models.SearchCriteria

	// Bind query parameters
	if err := c.ShouldBindQuery(&criteria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid search parameters",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if criteria.Limit <= 0 || criteria.Limit > 50 {
		criteria.Limit = 20
	}

	// For now, return mock search results
	// TODO: Implement actual database search
	mockRecipes := []models.RecipeData{
		{
			Title:       "豚キャベツ炒め",
			CookingTime: 10,
			Ingredients: []models.Ingredient{
				{Name: "豚こま肉", Amount: "200g"},
				{Name: "キャベツ", Amount: "1/4個"},
			},
			Steps:         []string{"材料を切る", "炒める", "味付けする"},
			LazinessScore: 9.0,
			Season:        "all",
			Tags:          []string{"簡単", "10分以内"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"recipes": mockRecipes,
			"total":   1,
			"limit":   criteria.Limit,
			"offset":  criteria.Offset,
		},
	})
}
