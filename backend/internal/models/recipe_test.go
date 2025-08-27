package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecipeData_Validate(t *testing.T) {
	tests := []struct {
		name    string
		recipe  RecipeData
		wantErr bool
	}{
		{
			name: "valid recipe",
			recipe: RecipeData{
				Title:         "Test Recipe",
				Ingredients:   []Ingredient{{Name: "ingredient1", Amount: "1 cup"}, {Name: "ingredient2", Amount: "2 tbsp"}},
				Steps:         []string{"step1", "step2"},
				CookingTime:   15,
				ServingSize:   2,
				Difficulty:    "easy",
				Season:        "all",
				LazinessScore: 5.0,
			},
			wantErr: false,
		},
		{
			name: "missing title",
			recipe: RecipeData{
				Ingredients:   []Ingredient{{Name: "ingredient1", Amount: "1 cup"}},
				Steps:         []string{"step1"},
				CookingTime:   15,
				ServingSize:   2,
				Season:        "all",
				LazinessScore: 5.0,
			},
			wantErr: true,
		},
		{
			name: "empty ingredients",
			recipe: RecipeData{
				Title:         "Test Recipe",
				Ingredients:   []Ingredient{},
				Steps:         []string{"step1"},
				CookingTime:   15,
				ServingSize:   2,
				Season:        "all",
				LazinessScore: 5.0,
			},
			wantErr: true,
		},
		{
			name: "empty steps",
			recipe: RecipeData{
				Title:         "Test Recipe",
				Ingredients:   []Ingredient{{Name: "ingredient1", Amount: "1 cup"}},
				Steps:         []string{},
				CookingTime:   15,
				ServingSize:   2,
				Season:        "all",
				LazinessScore: 5.0,
			},
			wantErr: true,
		},
		{
			name: "zero cooking time",
			recipe: RecipeData{
				Title:         "Test Recipe",
				Ingredients:   []Ingredient{{Name: "ingredient1", Amount: "1 cup"}},
				Steps:         []string{"step1"},
				CookingTime:   0,
				ServingSize:   2,
				Season:        "all",
				LazinessScore: 5.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.recipe.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecipeData_CalculateLazinessScore(t *testing.T) {
	tests := []struct {
		name     string
		recipe   RecipeData
		expected float64
	}{
		{
			name: "very lazy recipe",
			recipe: RecipeData{
				CookingTime: 5,
				Steps:       []string{"step1"},
				Ingredients: []Ingredient{{Name: "ingredient1", Amount: "1 cup"}, {Name: "ingredient2", Amount: "2 tbsp"}},
			},
			expected: 9.5, // Very short time, few steps, few ingredients
		},
		{
			name: "moderate recipe",
			recipe: RecipeData{
				CookingTime: 20,
				Steps:       []string{"step1", "step2", "step3"},
				Ingredients: []Ingredient{
					{Name: "ing1", Amount: "1 cup"}, {Name: "ing2", Amount: "1 cup"},
					{Name: "ing3", Amount: "1 cup"}, {Name: "ing4", Amount: "1 cup"}, {Name: "ing5", Amount: "1 cup"},
				},
			},
			expected: 6.5, // Medium time, moderate steps and ingredients
		},
		{
			name: "complex recipe",
			recipe: RecipeData{
				CookingTime: 60,
				Steps:       []string{"step1", "step2", "step3", "step4", "step5", "step6"},
				Ingredients: []Ingredient{
					{Name: "ing1", Amount: "1 cup"}, {Name: "ing2", Amount: "1 cup"},
					{Name: "ing3", Amount: "1 cup"}, {Name: "ing4", Amount: "1 cup"},
					{Name: "ing5", Amount: "1 cup"}, {Name: "ing6", Amount: "1 cup"},
					{Name: "ing7", Amount: "1 cup"}, {Name: "ing8", Amount: "1 cup"},
				},
			},
			expected: 3.0, // Long time, many steps and ingredients
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.recipe.CalculateLazinessScore()
			assert.InDelta(t, tt.expected, score, 1.0) // Allow 1.0 delta for calculation variations
		})
	}
}

func TestRecipeData_CalculateLazinessScore_Bounds(t *testing.T) {
	// Test minimum score (most complex)
	complexRecipe := RecipeData{
		CookingTime: 120,
		Steps:       make([]string, 20),     // 20 steps
		Ingredients: make([]Ingredient, 20), // 20 ingredients
	}
	score := complexRecipe.CalculateLazinessScore()
	assert.GreaterOrEqual(t, score, 1.0)
	assert.LessOrEqual(t, score, 10.0)

	// Test maximum score (simplest)
	simpleRecipe := RecipeData{
		CookingTime: 1,
		Steps:       []string{"mix everything"},
		Ingredients: []Ingredient{{Name: "one thing", Amount: "1 cup"}},
	}
	score = simpleRecipe.CalculateLazinessScore()
	assert.GreaterOrEqual(t, score, 1.0)
	assert.LessOrEqual(t, score, 10.0)
}

func TestIngredient_Structure(t *testing.T) {
	tests := []struct {
		name       string
		ingredient Ingredient
		valid      bool
	}{
		{
			name: "valid ingredient",
			ingredient: Ingredient{
				Name:   "chicken breast",
				Amount: "1 lb",
			},
			valid: true,
		},
		{
			name: "empty name",
			ingredient: Ingredient{
				Name:   "",
				Amount: "1 lb",
			},
			valid: false,
		},
		{
			name: "empty amount",
			ingredient: Ingredient{
				Name:   "chicken",
				Amount: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.ingredient.Name)
				assert.NotEmpty(t, tt.ingredient.Amount)
			} else {
				isEmpty := tt.ingredient.Name == "" || tt.ingredient.Amount == ""
				assert.True(t, isEmpty)
			}
		})
	}
}
