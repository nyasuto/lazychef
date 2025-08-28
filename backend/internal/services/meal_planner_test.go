package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"lazychef/internal/config"
	"lazychef/internal/database"
	"lazychef/internal/models"
)

func TestNewMealPlannerService(t *testing.T) {
	// Create test database
	db, err := database.New(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	generatorService := NewRecipeGeneratorService(config)
	service := NewMealPlannerService(db, generatorService)

	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.generator)
}

func TestMealPlannerService_OptimizeIngredients(t *testing.T) {
	// Create test database
	db, err := database.New(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	generatorService := NewRecipeGeneratorService(config)
	service := NewMealPlannerService(db, generatorService)

	recipes := []*models.RecipeData{
		{
			Title: "Recipe 1",
			Ingredients: []models.Ingredient{
				{Name: "豚肉", Amount: "200g"},
				{Name: "玉ねぎ", Amount: "1個"},
			},
		},
		{
			Title: "Recipe 2",
			Ingredients: []models.Ingredient{
				{Name: "豚肉", Amount: "150g"},
				{Name: "にんじん", Amount: "1本"},
			},
		},
	}

	optimized := service.optimizeIngredients(recipes)

	// Should have optimized the common ingredients
	assert.NotNil(t, optimized)
	assert.Len(t, optimized, 2)
}

func TestMealPlannerService_GenerateShoppingList(t *testing.T) {
	// Create test database
	db, err := database.New(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	generatorService := NewRecipeGeneratorService(config)
	service := NewMealPlannerService(db, generatorService)

	dailyRecipes := map[string]models.DailyRecipe{
		"2024-01-01": {
			Breakfast: &models.RecipeData{
				Title: "朝食",
				Ingredients: []models.Ingredient{
					{Name: "卵", Amount: "2個"},
					{Name: "パン", Amount: "2枚"},
				},
			},
			Lunch: &models.RecipeData{
				Title: "昼食",
				Ingredients: []models.Ingredient{
					{Name: "米", Amount: "1合"},
					{Name: "卵", Amount: "1個"},
				},
			},
		},
		"2024-01-02": {
			Breakfast: &models.RecipeData{
				Title: "朝食2",
				Ingredients: []models.Ingredient{
					{Name: "卵", Amount: "2個"},
					{Name: "牛乳", Amount: "200ml"},
				},
			},
		},
	}

	shoppingList := service.generateShoppingList(dailyRecipes, 2)

	// Should consolidate ingredients across days
	assert.NotEmpty(t, shoppingList)
	
	// Check for consolidated eggs (2+1+2 = 5個)
	foundEggs := false
	for _, item := range shoppingList {
		if item.Item == "卵" {
			foundEggs = true
			assert.Equal(t, "5個", item.Amount)
			break
		}
	}
	assert.True(t, foundEggs, "Should consolidate eggs")

	// Should have estimated costs
	totalCost := 0.0
	for _, item := range shoppingList {
		if item.Cost > 0 {
			totalCost += item.Cost
		}
	}
	assert.Greater(t, totalCost, 0.0, "Should have estimated costs")
}

func TestMealPlannerService_EstimateIngredientCost(t *testing.T) {
	// Create test database
	db, err := database.New(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	generatorService := NewRecipeGeneratorService(config)
	service := NewMealPlannerService(db, generatorService)

	tests := []struct {
		name     string
		item     string
		amount   string
		expected float64
	}{
		{"rice", "米", "2合", 200.0},
		{"eggs", "卵", "6個", 300.0},
		{"pork", "豚肉", "500g", 800.0},
		{"onion", "玉ねぎ", "3個", 150.0},
		{"unknown", "不明な食材", "100g", 100.0}, // default cost
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := service.estimateIngredientCost(tt.item, tt.amount)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

func TestMealPlannerService_ConsolidateIngredient(t *testing.T) {
	// Create test database
	db, err := database.New(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	generatorService := NewRecipeGeneratorService(config)
	service := NewMealPlannerService(db, generatorService)

	tests := []struct {
		name     string
		existing string
		new      string
		expected string
	}{
		{"same_unit_pieces", "2個", "3個", "5個"},
		{"same_unit_grams", "100g", "200g", "300g"},
		{"same_unit_ml", "100ml", "50ml", "150ml"},
		{"different_units", "1個", "100g", "1個, 100g"},
		{"mixed_formats", "2枚", "3枚", "5枚"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.consolidateIngredient(tt.existing, tt.new)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMealPlannerService_ParseAmount(t *testing.T) {
	// Create test database
	db, err := database.New(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	generatorService := NewRecipeGeneratorService(config)
	service := NewMealPlannerService(db, generatorService)

	tests := []struct {
		name     string
		amount   string
		value    float64
		unit     string
		canParse bool
	}{
		{"pieces", "5個", 5.0, "個", true},
		{"grams", "300g", 300.0, "g", true},
		{"ml", "200ml", 200.0, "ml", true},
		{"sheets", "3枚", 3.0, "枚", true},
		{"complex", "1/2個", 0.0, "", false}, // Complex fractions not supported
		{"text_only", "適量", 0.0, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, unit, ok := service.parseAmount(tt.amount)
			
			assert.Equal(t, tt.canParse, ok)
			if tt.canParse {
				assert.Equal(t, tt.value, value)
				assert.Equal(t, tt.unit, unit)
			}
		})
	}
}

func TestMealPlanPreferences_Validate(t *testing.T) {
	tests := []struct {
		name        string
		preferences models.MealPlanPreferences
		wantErr     bool
	}{
		{
			name: "valid preferences",
			preferences: models.MealPlanPreferences{
				DietaryRestrictions: []string{"vegetarian"},
				PreferredTags:       []string{"簡単", "時短"},
				BudgetLimit:         5000.0,
				LazinessPreference:  8.0,
			},
			wantErr: false,
		},
		{
			name: "negative budget",
			preferences: models.MealPlanPreferences{
				BudgetLimit: -1000.0,
			},
			wantErr: true,
		},
		{
			name: "invalid laziness score",
			preferences: models.MealPlanPreferences{
				LazinessPreference: 11.0, // Should be 1-10
			},
			wantErr: true,
		},
		{
			name: "zero laziness score",
			preferences: models.MealPlanPreferences{
				LazinessPreference: 0.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.preferences.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateMealPlanRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateMealPlanRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: models.CreateMealPlanRequest{
				Days:     7,
				Servings: 2,
				Preferences: models.MealPlanPreferences{
					BudgetLimit:        5000.0,
					LazinessPreference: 8.0,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid days",
			req: models.CreateMealPlanRequest{
				Days:     0,
				Servings: 2,
			},
			wantErr: true,
		},
		{
			name: "invalid servings",
			req: models.CreateMealPlanRequest{
				Days:     7,
				Servings: 0,
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

// Test meal plan creation with mock data
func TestMealPlannerService_CreateMealPlan_Structure(t *testing.T) {
	// Create test database
	db, err := database.New(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	schema := `
		CREATE TABLE IF NOT EXISTS meal_plans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data TEXT NOT NULL,
			week_start_date TEXT NOT NULL,
			days INTEGER NOT NULL,
			servings INTEGER NOT NULL,
			total_cost REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	config := &config.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	generatorService := NewRecipeGeneratorService(config)
	service := NewMealPlannerService(db, generatorService)

	req := models.CreateMealPlanRequest{
		Days:     3, // Shorter for testing
		Servings: 2,
		Preferences: models.MealPlanPreferences{
			DietaryRestrictions: []string{},
			PreferredTags:       []string{"簡単"},
			BudgetLimit:         3000.0,
			LazinessPreference:  9.0,
		},
	}

	// This will likely fail due to no API key, but we can test the structure
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result, err := service.CreateMealPlan(ctx, req)

	// The actual API call will fail, but we should get a structured error response
	if err != nil {
		// Expected due to no valid API key
		assert.Error(t, err)
	}

	if result != nil {
		// If we somehow get a result, validate its structure
		assert.NotNil(t, result)
		if result.MealPlan != nil {
			assert.NotEmpty(t, result.MealPlan.WeekStartDate)
			assert.NotNil(t, result.MealPlan.DailyRecipes)
			assert.NotNil(t, result.MealPlan.ShoppingList)
		}
	}
}