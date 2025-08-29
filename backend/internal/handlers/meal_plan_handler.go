package handlers

import (
	"lazychef/internal/models"
	"lazychef/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MealPlanHandler handles meal plan related requests
type MealPlanHandler struct {
	planner *services.MealPlannerService
}

// NewMealPlanHandler creates a new meal plan handler
func NewMealPlanHandler(planner *services.MealPlannerService) *MealPlanHandler {
	return &MealPlanHandler{
		planner: planner,
	}
}

// CreateMealPlan handles POST /api/meal-plans/create
func (h *MealPlanHandler) CreateMealPlan(c *gin.Context) {
	var req models.CreateMealPlanRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set defaults if not provided
	if req.StartDate == "" {
		req.StartDate = time.Now().Format("2006-01-02")
	}

	if req.Preferences.MaxCookingTime == 0 {
		req.Preferences.MaxCookingTime = 15
	}

	// Create meal plan
	mealPlan, err := h.planner.CreateWeeklyPlan(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create meal plan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mealPlan,
	})
}

// GetMealPlan handles GET /api/meal-plans/:id
func (h *MealPlanHandler) GetMealPlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meal plan ID",
		})
		return
	}

	mealPlan, err := h.planner.GetMealPlan(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Meal plan not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mealPlan,
	})
}

// ListMealPlans handles GET /api/meal-plans
func (h *MealPlanHandler) ListMealPlans(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	mealPlans, err := h.planner.ListMealPlans(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list meal plans",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"meal_plans": mealPlans,
			"limit":      limit,
			"offset":     offset,
		},
	})
}

// GenerateShoppingList handles POST /api/meal-plans/shopping-list
func (h *MealPlanHandler) GenerateShoppingList(c *gin.Context) {
	var req struct {
		RecipeIDs []int `json:"recipe_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate recipe IDs
	if len(req.RecipeIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No recipe IDs provided",
		})
		return
	}

	// Generate shopping list from recipe IDs
	shoppingList, err := h.planner.GenerateShoppingListFromRecipeIDs(req.RecipeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate shopping list",
			"details": err.Error(),
		})
		return
	}

	// Calculate total estimated cost
	totalCost := 0
	for range shoppingList {
		// Simple cost estimation: 200 yen per unique item
		totalCost += 200
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"shopping_list": shoppingList,
			"total_cost":    totalCost,
			"recipe_count":  len(req.RecipeIDs),
		},
	})
}
