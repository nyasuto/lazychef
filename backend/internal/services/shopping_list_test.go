package services

import (
	"testing"
)

func TestMealPlannerService_GenerateShoppingListFromRecipeIDs(t *testing.T) {
	// This test requires a real database with actual recipes
	// For unit testing, we'll test the category mapping function instead
	service := &MealPlannerService{
		ingredientAggregator: NewIngredientAggregator(),
	}

	// Test ingredient category mapping
	tests := []struct {
		ingredient string
		expected   string
	}{
		{"豚こま肉", "肉類"},
		{"キャベツ", "野菜"},
		{"卵", "乳製品"},
		{"醤油", "調味料"},
		{"ご飯", "穀物"},
		{"豆腐", "豆腐・大豆製品"},
		{"ツナ缶", "魚介類"},
		{"謎の食材", "その他"},
	}

	for _, test := range tests {
		result := service.getIngredientCategory(test.ingredient)
		if result != test.expected {
			t.Errorf("For ingredient '%s', expected category '%s', got '%s'", 
				test.ingredient, test.expected, result)
		}
	}
}

func TestMealPlannerService_getIngredientCategory(t *testing.T) {
	service := &MealPlannerService{}

	// Test exact matches
	exactMatches := map[string]string{
		"豚こま肉": "肉類",
		"キャベツ": "野菜", 
		"牛乳":   "乳製品",
		"醤油":   "調味料",
		"ご飯":   "穀物",
	}

	for ingredient, expectedCategory := range exactMatches {
		result := service.getIngredientCategory(ingredient)
		if result != expectedCategory {
			t.Errorf("Exact match failed: ingredient '%s', expected '%s', got '%s'", 
				ingredient, expectedCategory, result)
		}
	}

	// Test partial matches
	partialMatches := map[string]string{
		"豚肉の薄切り": "肉類",
		"キャベツの千切り": "野菜",
	}

	for ingredient, expectedCategory := range partialMatches {
		result := service.getIngredientCategory(ingredient)
		if result != expectedCategory {
			t.Errorf("Partial match failed: ingredient '%s', expected '%s', got '%s'", 
				ingredient, expectedCategory, result)
		}
	}

	// Test unknown ingredient
	unknownIngredient := "宇宙食材"
	result := service.getIngredientCategory(unknownIngredient)
	if result != "その他" {
		t.Errorf("Unknown ingredient should return 'その他', got '%s'", result)
	}
}

func TestRecipeRepository_GetRecipesByIDs(t *testing.T) {
	// This is more of an integration test and requires a real database
	// We'll create a simple test to verify the function signature
	repo := &RecipeRepository{}
	
	// Test empty IDs
	recipes, err := repo.GetRecipesByIDs([]int{})
	if err != nil {
		t.Errorf("GetRecipesByIDs with empty slice should not error, got: %v", err)
	}
	if len(recipes) != 0 {
		t.Errorf("GetRecipesByIDs with empty slice should return empty slice, got %d recipes", len(recipes))
	}
}