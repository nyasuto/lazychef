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
	// Convert meal plan data to JSON for storage
	weekDataJSON, err := json.Marshal(plan.WeekData)
	if err != nil {
		return fmt.Errorf("failed to marshal meal plan data: %w", err)
	}

	query := `
		INSERT INTO meal_plans (week_data)
		VALUES (?)
	`

	// Execute the database query
	if err := s.db.Execute(query, string(weekDataJSON)); err != nil {
		return fmt.Errorf("failed to execute meal plan insert: %w", err)
	}

	// Get the last inserted ID
	id, err := s.db.GetLastInsertID()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	// Set the ID on the plan object
	plan.ID = int(id)

	return nil
}

// GetMealPlan retrieves a meal plan by ID
func (s *MealPlannerService) GetMealPlan(id int) (*models.MealPlan, error) {
	query := `
		SELECT id, week_data, created_at
		FROM meal_plans
		WHERE id = ?
	`

	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query meal plan: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// In a real application, we would log this error
			_ = closeErr
		}
	}()

	if !rows.Next() {
		return nil, fmt.Errorf("meal plan with id %d not found", id)
	}

	var mealPlan models.MealPlan
	var weekDataJSON string
	var createdAt string

	if err := rows.Scan(&mealPlan.ID, &weekDataJSON, &createdAt); err != nil {
		return nil, fmt.Errorf("failed to scan meal plan: %w", err)
	}

	// Parse JSON data
	if err := json.Unmarshal([]byte(weekDataJSON), &mealPlan.WeekData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meal plan data: %w", err)
	}

	return &mealPlan, nil
}

// ListMealPlans lists meal plans with pagination
func (s *MealPlannerService) ListMealPlans(limit, offset int) ([]*models.MealPlan, error) {
	// Set reasonable limits
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, week_data, created_at
		FROM meal_plans
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query meal plans: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// In a real application, we would log this error
			_ = closeErr
		}
	}()

	plans := make([]*models.MealPlan, 0, limit)

	for rows.Next() {
		var mealPlan models.MealPlan
		var weekDataJSON string
		var createdAt string

		if err := rows.Scan(&mealPlan.ID, &weekDataJSON, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan meal plan: %w", err)
		}

		// Parse JSON data
		if err := json.Unmarshal([]byte(weekDataJSON), &mealPlan.WeekData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meal plan data: %w", err)
		}

		plans = append(plans, &mealPlan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return plans, nil
}
