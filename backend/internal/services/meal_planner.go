package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"lazychef/internal/database"
	"lazychef/internal/models"
)

// MealPlannerService handles meal plan creation and management
type MealPlannerService struct {
	db                   *database.Database
	generator            *RecipeGeneratorService
	ingredientAggregator *IngredientAggregator
	recipeRepo           *RecipeRepository
}

// NewMealPlannerService creates a new meal planner service
func NewMealPlannerService(db *database.Database, generator *RecipeGeneratorService) *MealPlannerService {
	return &MealPlannerService{
		db:                   db,
		generator:            generator,
		ingredientAggregator: NewIngredientAggregator(),
		recipeRepo:           NewRecipeRepository(db),
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
	// Map to collect quantities for each ingredient
	ingredientQuantitiesMap := make(map[string][]*IngredientQuantity)

	// Collect all ingredient quantities
	for _, recipe := range recipes {
		for _, ingredient := range recipe.Ingredients {
			qty, err := s.ingredientAggregator.ParseQuantity(ingredient.Amount)
			if err != nil {
				// If parsing fails, use "適量"
				qty = &IngredientQuantity{Amount: 0, Unit: "適量"}
			}

			ingredientQuantitiesMap[ingredient.Name] = append(
				ingredientQuantitiesMap[ingredient.Name],
				qty,
			)
		}
	}

	// Aggregate quantities for each ingredient
	shoppingList := make([]models.ShoppingItem, 0, len(ingredientQuantitiesMap))
	for ingredientName, quantities := range ingredientQuantitiesMap {
		aggregatedQty, err := s.ingredientAggregator.AggregateQuantities(quantities)
		if err != nil {
			// If aggregation fails, use "適量"
			aggregatedQty = &IngredientQuantity{Amount: 0, Unit: "適量"}
		}

		// Format the aggregated quantity
		amountStr := s.ingredientAggregator.FormatQuantity(aggregatedQty)

		shoppingList = append(shoppingList, models.ShoppingItem{
			Item:   ingredientName,
			Amount: amountStr,
		})
	}

	return shoppingList
}

// estimateTotalCost estimates the total cost of shopping
func (s *MealPlannerService) estimateTotalCost(items []models.ShoppingItem) float64 {
	// Simple estimation: 200 yen per item average
	return float64(len(items)) * 200
}

// GenerateShoppingListFromRecipeIDs generates shopping list from recipe IDs
func (s *MealPlannerService) GenerateShoppingListFromRecipeIDs(recipeIDs []int) ([]models.ShoppingItem, error) {
	// Get recipes by IDs
	recipes, err := s.recipeRepo.GetRecipesByIDs(recipeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}

	// Convert Recipe objects to RecipeData objects
	recipeDataList := make([]models.RecipeData, 0, len(recipes))
	for _, recipe := range recipes {
		recipeDataList = append(recipeDataList, recipe.Data)
	}

	// Generate shopping list using existing logic
	shoppingList := s.createShoppingList(recipeDataList)

	// Add categories to shopping items
	for i := range shoppingList {
		shoppingList[i].Category = s.getIngredientCategory(shoppingList[i].Item)
	}

	return shoppingList, nil
}

// getIngredientCategory determines the category of an ingredient
func (s *MealPlannerService) getIngredientCategory(ingredient string) string {
	// Define category mapping
	categoryMap := map[string]string{
		// 野菜
		"キャベツ": "野菜", "レタス": "野菜", "トマト": "野菜", "きゅうり": "野菜",
		"玉ねぎ": "野菜", "にんじん": "野菜", "じゃがいも": "野菜", "大根": "野菜",
		"ブロッコリー": "野菜", "ほうれん草": "野菜", "小松菜": "野菜", "白菜": "野菜",
		"なす": "野菜", "ピーマン": "野菜", "もやし": "野菜", "ねぎ": "野菜", "長ねぎ": "野菜",

		// 肉類
		"豚肉": "肉類", "豚こま肉": "肉類", "豚ロース": "肉類", "豚バラ肉": "肉類",
		"鶏肉": "肉類", "鶏もも肉": "肉類", "鶏むね肉": "肉類", "鶏ひき肉": "肉類",
		"牛肉": "肉類", "合いびき肉": "肉類", "ひき肉": "肉類",

		// 魚介類
		"サーモン": "魚介類", "まぐろ": "魚介類", "サバ": "魚介類", "アジ": "魚介類",
		"エビ": "魚介類", "イカ": "魚介類", "ツナ缶": "魚介類",

		// 乳製品・卵
		"牛乳": "乳製品", "チーズ": "乳製品", "バター": "乳製品", "ヨーグルト": "乳製品",
		"卵": "乳製品",

		// 穀物・パン類
		"米": "穀物", "ご飯": "穀物", "パン": "穀物", "食パン": "穀物",
		"うどん": "穀物", "そば": "穀物", "パスタ": "穀物", "小麦粉": "穀物",

		// 調味料
		"醤油": "調味料", "味噌": "調味料", "塩": "調味料", "砂糖": "調味料",
		"酢": "調味料", "みりん": "調味料", "酒": "調味料", "料理酒": "調味料",
		"ごま油": "調味料", "サラダ油": "調味料", "オリーブオイル": "調味料",
		"こしょう": "調味料", "胡椒": "調味料", "マヨネーズ": "調味料", "ケチャップ": "調味料",
		"ソース": "調味料", "だしの素": "調味料", "コンソメ": "調味料", "鶏がらスープの素": "調味料",

		// 豆腐・大豆製品
		"豆腐": "豆腐・大豆製品", "厚揚げ": "豆腐・大豆製品", "油揚げ": "豆腐・大豆製品", "納豆": "豆腐・大豆製品",
	}

	// Check for exact match first
	if category, exists := categoryMap[ingredient]; exists {
		return category
	}

	// Check for partial matches
	for key, category := range categoryMap {
		if strings.Contains(ingredient, key) || strings.Contains(key, ingredient) {
			return category
		}
	}

	// Default category
	return "その他"
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
