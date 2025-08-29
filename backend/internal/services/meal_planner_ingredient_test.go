package services

import (
	"testing"

	"lazychef/internal/models"
)

func TestMealPlannerService_createShoppingList_Aggregation(t *testing.T) {
	service := &MealPlannerService{
		ingredientAggregator: NewIngredientAggregator(),
	}

	// Test recipes with overlapping ingredients
	recipes := []models.RecipeData{
		{
			Title: "Recipe 1",
			Ingredients: []models.Ingredient{
				{Name: "卵", Amount: "2個"},
				{Name: "牛乳", Amount: "200ml"},
				{Name: "小麦粉", Amount: "100g"},
			},
		},
		{
			Title: "Recipe 2",
			Ingredients: []models.Ingredient{
				{Name: "卵", Amount: "3個"},
				{Name: "牛乳", Amount: "150ml"},
				{Name: "砂糖", Amount: "50g"},
			},
		},
	}

	result := service.createShoppingList(recipes)

	// Check results
	ingredientMap := make(map[string]string)
	for _, item := range result {
		ingredientMap[item.Item] = item.Amount
	}

	// Verify aggregation results
	tests := []struct {
		ingredient string
		expected   string
	}{
		{"卵", "5個"},
		{"牛乳", "350ml"},
		{"小麦粉", "100g"},
		{"砂糖", "50g"},
	}

	for _, test := range tests {
		amount, exists := ingredientMap[test.ingredient]
		if !exists {
			t.Errorf("Expected ingredient '%s' not found in shopping list", test.ingredient)
			continue
		}
		if amount != test.expected {
			t.Errorf("For ingredient '%s', expected '%s', got '%s'",
				test.ingredient, test.expected, amount)
		}
	}

	// Check total number of items
	expectedItems := 4
	if len(result) != expectedItems {
		t.Errorf("Expected %d items in shopping list, got %d", expectedItems, len(result))
	}
}

func TestMealPlannerService_createShoppingList_DifferentUnits(t *testing.T) {
	service := &MealPlannerService{
		ingredientAggregator: NewIngredientAggregator(),
	}

	// Test recipes with different units for same ingredient
	recipes := []models.RecipeData{
		{
			Title: "Recipe 1",
			Ingredients: []models.Ingredient{
				{Name: "砂糖", Amount: "100g"},
			},
		},
		{
			Title: "Recipe 2",
			Ingredients: []models.Ingredient{
				{Name: "砂糖", Amount: "1kg"},
			},
		},
	}

	result := service.createShoppingList(recipes)

	// Find sugar in the shopping list
	var sugarAmount string
	for _, item := range result {
		if item.Item == "砂糖" {
			sugarAmount = item.Amount
			break
		}
	}

	expected := "1.1kg"
	if sugarAmount != expected {
		t.Errorf("For aggregated sugar, expected '%s', got '%s'", expected, sugarAmount)
	}
}

func TestMealPlannerService_createShoppingList_WithTekiRyou(t *testing.T) {
	service := &MealPlannerService{
		ingredientAggregator: NewIngredientAggregator(),
	}

	// Test recipes with "適量" ingredients
	recipes := []models.RecipeData{
		{
			Title: "Recipe 1",
			Ingredients: []models.Ingredient{
				{Name: "塩", Amount: "2g"},
				{Name: "胡椒", Amount: "適量"},
			},
		},
		{
			Title: "Recipe 2",
			Ingredients: []models.Ingredient{
				{Name: "塩", Amount: "3g"},
				{Name: "胡椒", Amount: "少々"},
			},
		},
	}

	result := service.createShoppingList(recipes)

	ingredientMap := make(map[string]string)
	for _, item := range result {
		ingredientMap[item.Item] = item.Amount
	}

	// Check results
	tests := []struct {
		ingredient string
		expected   string
	}{
		{"塩", "5g"},  // Should aggregate normally
		{"胡椒", "適量"}, // Should become "適量" when mixed with "適量"/"少々"
	}

	for _, test := range tests {
		amount, exists := ingredientMap[test.ingredient]
		if !exists {
			t.Errorf("Expected ingredient '%s' not found in shopping list", test.ingredient)
			continue
		}
		if amount != test.expected {
			t.Errorf("For ingredient '%s', expected '%s', got '%s'",
				test.ingredient, test.expected, amount)
		}
	}
}
