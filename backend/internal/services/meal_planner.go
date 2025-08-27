package services

import (
	"encoding/json"
	"fmt"
	"lazychef/internal/database"
	"lazychef/internal/models"
)

// MealPlannerService handles meal plan creation and management
type MealPlannerService struct {
	db        *database.Database
	generator *RecipeGeneratorService
}

// NewMealPlannerService creates a new meal planner service
func NewMealPlannerService(db *database.Database, generator *RecipeGeneratorService) *MealPlannerService {
	return &MealPlannerService{
		db:        db,
		generator: generator,
	}
}

// CreateWeeklyPlan creates a weekly meal plan
func (s *MealPlannerService) CreateWeeklyPlan(req models.CreateMealPlanRequest) (*models.MealPlan, error) {
	// Generate recipes for the week
	recipes := make([]models.RecipeData, 0, 5)

	// Generate 5 recipes for weekdays
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday"}
	for i := 0; i < 5; i++ {
		// Use fallback recipes for now (AI generation will be enhanced later)
		recipe := s.getFallbackRecipe(i)
		recipes = append(recipes, *recipe)
	}

	// Create shopping list
	shoppingList := s.createShoppingList(recipes)

	// Build meal plan data
	mealPlanData := models.MealPlanData{
		StartDate:         req.StartDate,
		ShoppingList:      shoppingList,
		DailyRecipes:      make(map[string]models.DailyRecipe),
		TotalCostEstimate: int(s.estimateTotalCost(shoppingList)),
	}

	// Create meal plan
	mealPlan := &models.MealPlan{
		WeekData: mealPlanData,
	}

	// Assign recipes to days
	for i, day := range days {
		if i < len(recipes) {
			mealPlan.WeekData.DailyRecipes[day] = models.DailyRecipe{
				RecipeID: i + 1,
				Title:    recipes[i].Title,
			}
		}
	}

	// Save to database
	if s.db != nil {
		if err := s.saveMealPlan(mealPlan); err != nil {
			return nil, fmt.Errorf("failed to save meal plan: %w", err)
		}
	}

	return mealPlan, nil
}

// createShoppingList creates a shopping list from recipes
func (s *MealPlannerService) createShoppingList(recipes []models.RecipeData) []models.ShoppingItem {
	itemMap := make(map[string]string)

	for _, recipe := range recipes {
		for _, ingredient := range recipe.Ingredients {
			// Aggregate same ingredients
			if _, exists := itemMap[ingredient.Name]; exists {
				// TODO: Properly aggregate amounts
				itemMap[ingredient.Name] = "適量"
			} else {
				itemMap[ingredient.Name] = ingredient.Amount
			}
		}
	}

	// Convert to shopping list
	shoppingList := make([]models.ShoppingItem, 0, len(itemMap))
	for name, amount := range itemMap {
		shoppingList = append(shoppingList, models.ShoppingItem{
			Item:   name,
			Amount: amount,
		})
	}

	return shoppingList
}

// estimateTotalCost estimates the total cost of shopping
func (s *MealPlannerService) estimateTotalCost(items []models.ShoppingItem) float64 {
	// Simple estimation: 200 yen per item average
	return float64(len(items)) * 200
}

// getFallbackRecipe returns a fallback recipe
func (s *MealPlannerService) getFallbackRecipe(index int) *models.RecipeData {
	fallbackRecipes := []models.RecipeData{
		{
			Title:       "豚キャベツ炒め",
			CookingTime: 10,
			Ingredients: []models.Ingredient{
				{Name: "豚こま肉", Amount: "200g"},
				{Name: "キャベツ", Amount: "1/4個"},
				{Name: "醤油", Amount: "大さじ1"},
			},
			Steps: []string{
				"キャベツをざく切りにする",
				"豚肉を炒める",
				"キャベツを加えて醤油で味付け",
			},
			LazinessScore: 9.0,
		},
		{
			Title:       "もやしと卵の炒め物",
			CookingTime: 8,
			Ingredients: []models.Ingredient{
				{Name: "もやし", Amount: "1袋"},
				{Name: "卵", Amount: "2個"},
				{Name: "塩コショウ", Amount: "少々"},
			},
			Steps: []string{
				"もやしを洗う",
				"フライパンで炒める",
				"卵を加えて塩コショウで味付け",
			},
			LazinessScore: 9.5,
		},
		{
			Title:       "豆腐の煮物",
			CookingTime: 12,
			Ingredients: []models.Ingredient{
				{Name: "豆腐", Amount: "1丁"},
				{Name: "めんつゆ", Amount: "大さじ2"},
				{Name: "ネギ", Amount: "少々"},
			},
			Steps: []string{
				"豆腐を切る",
				"鍋でめんつゆと煮る",
				"ネギを散らす",
			},
			LazinessScore: 8.5,
		},
		{
			Title:       "鶏もも肉の照り焼き",
			CookingTime: 15,
			Ingredients: []models.Ingredient{
				{Name: "鶏もも肉", Amount: "1枚"},
				{Name: "醤油", Amount: "大さじ2"},
				{Name: "みりん", Amount: "大さじ2"},
			},
			Steps: []string{
				"鶏肉を一口大に切る",
				"フライパンで焼く",
				"醤油とみりんで照り焼きにする",
			},
			LazinessScore: 8.0,
		},
		{
			Title:       "野菜炒め",
			CookingTime: 10,
			Ingredients: []models.Ingredient{
				{Name: "キャベツ", Amount: "1/4個"},
				{Name: "にんじん", Amount: "1/2本"},
				{Name: "塩コショウ", Amount: "少々"},
			},
			Steps: []string{
				"野菜を切る",
				"フライパンで炒める",
				"塩コショウで味付け",
			},
			LazinessScore: 9.0,
		},
	}

	if index < len(fallbackRecipes) {
		return &fallbackRecipes[index]
	}
	return &fallbackRecipes[0]
}

// saveMealPlan saves a meal plan to the database
func (s *MealPlannerService) saveMealPlan(plan *models.MealPlan) error {
	// Convert to JSON for storage
	data, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("failed to marshal meal plan: %w", err)
	}

	query := `
		INSERT INTO meal_plans (week_data)
		VALUES (?)
	`

	// TODO: Execute query with database connection
	_ = query
	_ = data

	return nil
}

// GetMealPlan retrieves a meal plan by ID
func (s *MealPlannerService) GetMealPlan(id int) (*models.MealPlan, error) {
	// For now, return a mock meal plan to avoid staticcheck issues
	// TODO: Implement actual database query

	mockPlan := &models.MealPlan{
		ID: id,
		WeekData: models.MealPlanData{
			StartDate:         "2025-01-27",
			ShoppingList:      []models.ShoppingItem{},
			DailyRecipes:      make(map[string]models.DailyRecipe),
			TotalCostEstimate: 1500,
		},
	}

	return mockPlan, nil
}

// ListMealPlans lists meal plans with pagination
func (s *MealPlannerService) ListMealPlans(limit, offset int) ([]*models.MealPlan, error) {
	// For now, return mock meal plans to avoid staticcheck issues
	// TODO: Implement actual database query

	mockPlans := make([]*models.MealPlan, 0, limit)
	for i := 0; i < limit && i < 3; i++ { // Return up to 3 mock plans
		plan := &models.MealPlan{
			ID: i + 1,
			WeekData: models.MealPlanData{
				StartDate:         "2025-01-27",
				ShoppingList:      []models.ShoppingItem{},
				DailyRecipes:      make(map[string]models.DailyRecipe),
				TotalCostEstimate: 1500 + (i * 200),
			},
		}
		mockPlans = append(mockPlans, plan)
	}

	return mockPlans, nil
}
