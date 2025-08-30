package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"lazychef/internal/database"
	"lazychef/internal/models"
	"lazychef/internal/services"
)

// RecipeHandler handles recipe-related HTTP requests
type RecipeHandler struct {
	db                       *database.Database
	recipeRepository         *services.RecipeRepository
	generatorService         *services.RecipeGeneratorService
	enhancedGeneratorService *services.EnhancedRecipeGeneratorService
	ingredientMapper         *services.SimpleIngredientMapper
}

// NewRecipeHandler creates a new recipe handler
func NewRecipeHandler(db *database.Database, generatorService *services.RecipeGeneratorService, enhancedGeneratorService *services.EnhancedRecipeGeneratorService) *RecipeHandler {
	recipeRepository := services.NewRecipeRepository(db)
	ingredientMapper := services.NewSimpleIngredientMapper()
	return &RecipeHandler{
		db:                       db,
		recipeRepository:         recipeRepository,
		generatorService:         generatorService,
		enhancedGeneratorService: enhancedGeneratorService,
		ingredientMapper:         ingredientMapper,
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

	// Save the generated recipe to database
	recipe := &models.Recipe{
		Data:      *result.Recipe,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.recipeRepository.SaveRecipe(recipe); err != nil {
		// Log the error but don't fail the request - recipe generation was successful
		// In a production system, you might want to queue this for retry
		c.JSON(http.StatusOK, gin.H{
			"recipe":       result.Recipe,
			"metadata":     result.Metadata,
			"save_warning": "Recipe generated but failed to save to database",
			"save_error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recipe":    result.Recipe,
		"recipe_id": recipe.ID,
		"metadata":  result.Metadata,
		"saved":     true,
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

	// Save all generated recipes to database
	recipeIDs := make([]int, 0, len(result.Recipes))
	savedCount := 0
	var saveErrors []string

	for _, recipeData := range result.Recipes {
		recipe := &models.Recipe{
			Data:      recipeData,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := h.recipeRepository.SaveRecipe(recipe); err != nil {
			saveErrors = append(saveErrors, err.Error())
		} else {
			recipeIDs = append(recipeIDs, recipe.ID)
			savedCount++
		}
	}

	response := gin.H{
		"recipes":     result.Recipes,
		"metadata":    result.Metadata,
		"saved_count": savedCount,
		"total_count": len(result.Recipes),
	}

	if len(recipeIDs) > 0 {
		response["recipe_ids"] = recipeIDs
		response["saved"] = true
	}

	if len(saveErrors) > 0 {
		response["save_errors"] = saveErrors
		response["save_warning"] = "Some recipes failed to save to database"
	}

	c.JSON(http.StatusOK, response)
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
// SearchRecipes handles GET /api/recipes/search
func (h *RecipeHandler) SearchRecipes(c *gin.Context) {
	var criteria models.SearchCriteria

	// Debug: Log all query parameters
	fmt.Printf("[DEBUG] Raw query parameters: %+v\n", c.Request.URL.Query())

	// Bind query parameters
	if err := c.ShouldBindQuery(&criteria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid search parameters",
			"details": err.Error(),
		})
		return
	}

	// Debug: Log bound criteria
	fmt.Printf("[DEBUG] Bound criteria: %+v\n", criteria)

	// Handle comma-separated ingredients from frontend
	if ingredientsParam := c.Query("ingredients"); ingredientsParam != "" && len(criteria.Ingredients) == 0 {
		criteria.Ingredients = strings.Split(strings.TrimSpace(ingredientsParam), ",")
		// Clean up each ingredient
		for i, ingredient := range criteria.Ingredients {
			criteria.Ingredients[i] = strings.TrimSpace(ingredient)
		}
	}

	// Handle comma-separated tags from frontend
	if tagsParam := c.Query("tags"); tagsParam != "" && len(criteria.Tags) == 0 {
		criteria.Tags = strings.Split(strings.TrimSpace(tagsParam), ",")
		// Clean up each tag
		for i, tag := range criteria.Tags {
			criteria.Tags[i] = strings.TrimSpace(tag)
		}
	}

	// Debug: Log criteria after comma-separated processing
	fmt.Printf("[DEBUG] Processed criteria: %+v\n", criteria)

	// Set defaults and validate
	if criteria.Limit <= 0 || criteria.Limit > 100 {
		criteria.Limit = 20
	}
	if criteria.Offset < 0 {
		criteria.Offset = 0
	}

	// Handle page-based pagination (alternative to offset)
	if criteria.Page > 0 {
		criteria.Offset = (criteria.Page - 1) * criteria.Limit
	}

	// Build SQL query with conditions
	query := `
		SELECT id, data, created_at
		FROM recipes
		WHERE 1=1
	`
	args := []interface{}{}

	// Add search conditions
	if criteria.Query != "" {
		query += ` AND (
			title LIKE ? OR 
			json_extract(data, '$.ingredients') LIKE ?
		)`
		searchTerm := "%" + criteria.Query + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Handle multiple tags
	if len(criteria.Tags) > 0 {
		for _, tag := range criteria.Tags {
			query += ` AND json_extract(data, '$.tags') LIKE ?`
			args = append(args, "%"+tag+"%")
		}
	} else if criteria.Tag != "" { // Backward compatibility
		query += ` AND json_extract(data, '$.tags') LIKE ?`
		args = append(args, "%"+criteria.Tag+"%")
	}

	// Handle multiple ingredients - use synonym mapping for better matching
	if len(criteria.Ingredients) > 0 {
		// Expand ingredients using synonym mapper (Issue #87 fix)
		expandedIngredients := h.ingredientMapper.ExpandIngredientTerms(criteria.Ingredients)
		fmt.Printf("[DEBUG] Original ingredients: %v, Expanded: %v\n", criteria.Ingredients, expandedIngredients)

		if len(expandedIngredients) > 0 {
			// Build OR conditions for all expanded ingredients
			var orConditions []string
			for _, ingredient := range expandedIngredients {
				orConditions = append(orConditions, `EXISTS (
					SELECT 1 FROM json_each(json_extract(data, '$.ingredients'))
					WHERE json_extract(value, '$.name') = ?
				)`)
				args = append(args, ingredient)
			}
			query += ` AND (` + strings.Join(orConditions, " OR ") + `)`
		}
	} else if criteria.Ingredient != "" { // Backward compatibility
		// Also apply synonym mapping to single ingredient searches
		expandedIngredients := h.ingredientMapper.ExpandIngredientTerms([]string{criteria.Ingredient})
		fmt.Printf("[DEBUG] Single ingredient '%s' expanded to: %v\n", criteria.Ingredient, expandedIngredients)

		if len(expandedIngredients) > 0 {
			var orConditions []string
			for _, ingredient := range expandedIngredients {
				orConditions = append(orConditions, `EXISTS (
					SELECT 1 FROM json_each(json_extract(data, '$.ingredients'))
					WHERE json_extract(value, '$.name') = ?
				)`)
				args = append(args, ingredient)
			}
			query += ` AND (` + strings.Join(orConditions, " OR ") + `)`
		}
	}

	if criteria.MaxCookingTime > 0 {
		query += ` AND cooking_time <= ?`
		args = append(args, criteria.MaxCookingTime)
	}

	if criteria.MinLazinessScore > 0 {
		query += ` AND laziness_score >= ?`
		args = append(args, criteria.MinLazinessScore)
	}

	if criteria.Season != "" && criteria.Season != "all" {
		query += ` AND (season = ? OR season = 'all')`
		args = append(args, criteria.Season)
	}

	// Add ordering and pagination
	query += ` ORDER BY laziness_score DESC, created_at DESC LIMIT ? OFFSET ?`
	args = append(args, criteria.Limit, criteria.Offset)

	// Execute query
	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database query failed",
			"details": err.Error(),
		})
		return
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// In a real application, we would log this error
			_ = closeErr
		}
	}()

	// Parse results
	recipes := make([]models.RecipeData, 0, criteria.Limit)
	for rows.Next() {
		var id int
		var dataJSON string
		var createdAt string

		if err := rows.Scan(&id, &dataJSON, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan recipe data",
				"details": err.Error(),
			})
			return
		}

		// Parse JSON data
		var recipe models.RecipeData
		if err := json.Unmarshal([]byte(dataJSON), &recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to parse recipe JSON",
				"details": err.Error(),
			})
			return
		}

		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error during row iteration",
			"details": err.Error(),
		})
		return
	}

	// Get total count for pagination info (separate query for performance)
	countQuery := `SELECT COUNT(*) FROM recipes WHERE 1=1`
	countArgs := []interface{}{}

	// Re-build count query with same conditions (without LIMIT/OFFSET)
	if criteria.Query != "" {
		countQuery += ` AND (title LIKE ? OR json_extract(data, '$.ingredients') LIKE ?)`
		searchTerm := "%" + criteria.Query + "%"
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	// Handle multiple tags for count
	if len(criteria.Tags) > 0 {
		for _, tag := range criteria.Tags {
			countQuery += ` AND json_extract(data, '$.tags') LIKE ?`
			countArgs = append(countArgs, "%"+tag+"%")
		}
	} else if criteria.Tag != "" { // Backward compatibility
		countQuery += ` AND json_extract(data, '$.tags') LIKE ?`
		countArgs = append(countArgs, "%"+criteria.Tag+"%")
	}

	// Handle multiple ingredients for count - use synonym mapping for better matching
	if len(criteria.Ingredients) > 0 {
		// Use same expanded ingredients as main query
		expandedIngredients := h.ingredientMapper.ExpandIngredientTerms(criteria.Ingredients)
		if len(expandedIngredients) > 0 {
			var orConditions []string
			for _, ingredient := range expandedIngredients {
				orConditions = append(orConditions, `EXISTS (
					SELECT 1 FROM json_each(json_extract(data, '$.ingredients'))
					WHERE json_extract(value, '$.name') = ?
				)`)
				countArgs = append(countArgs, ingredient)
			}
			countQuery += ` AND (` + strings.Join(orConditions, " OR ") + `)`
		}
	} else if criteria.Ingredient != "" { // Backward compatibility
		// Use same expanded ingredients as main query
		expandedIngredients := h.ingredientMapper.ExpandIngredientTerms([]string{criteria.Ingredient})
		if len(expandedIngredients) > 0 {
			var orConditions []string
			for _, ingredient := range expandedIngredients {
				orConditions = append(orConditions, `EXISTS (
					SELECT 1 FROM json_each(json_extract(data, '$.ingredients'))
					WHERE json_extract(value, '$.name') = ?
				)`)
				countArgs = append(countArgs, ingredient)
			}
			countQuery += ` AND (` + strings.Join(orConditions, " OR ") + `)`
		}
	}
	if criteria.MaxCookingTime > 0 {
		countQuery += ` AND cooking_time <= ?`
		countArgs = append(countArgs, criteria.MaxCookingTime)
	}
	if criteria.MinLazinessScore > 0 {
		countQuery += ` AND laziness_score >= ?`
		countArgs = append(countArgs, criteria.MinLazinessScore)
	}
	if criteria.Season != "" && criteria.Season != "all" {
		countQuery += ` AND (season = ? OR season = 'all')`
		countArgs = append(countArgs, criteria.Season)
	}

	var total int
	if err := h.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		// If count fails, log but don't fail the whole request
		total = len(recipes)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"recipes": recipes,
			"total":   total,
			"limit":   criteria.Limit,
			"offset":  criteria.Offset,
			"page":    criteria.Page,
		},
	})
}

// GetIngredientCategories handles GET /api/recipes/ingredient-categories
// Returns available ingredient categories for UI dropdown
func (h *RecipeHandler) GetIngredientCategories(c *gin.Context) {
	categories := h.ingredientMapper.GetSupportedCategories()

	// Return both categories and their associated ingredients for debugging
	categoryDetails := make(map[string]interface{})
	for _, category := range categories {
		ingredients := h.ingredientMapper.GetCategoryIngredients(category)
		categoryDetails[category] = gin.H{
			"display_name": category,
			"ingredients":  ingredients,
			"count":        len(ingredients),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"categories": categories,
			"details":    categoryDetails,
		},
	})
}

// TestIngredientMapping handles GET /api/recipes/test-ingredient-mapping
// Test endpoint to verify ingredient synonym mapping is working
func (h *RecipeHandler) TestIngredientMapping(c *gin.Context) {
	searchTerm := c.Query("term")
	if searchTerm == "" {
		searchTerm = "鶏肉" // Default test term
	}

	expandedIngredients := h.ingredientMapper.ExpandIngredientTerms([]string{searchTerm})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"search_term":          searchTerm,
			"expanded_ingredients": expandedIngredients,
			"expansion_count":      len(expandedIngredients),
			"note":                 "This tests the ingredient synonym mapping system (Issue #87 fix)",
		},
	})
}
