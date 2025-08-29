package models

import (
	"encoding/json"
	"time"
)

// ShoppingItem represents an item in the shopping list
type ShoppingItem struct {
	Item     string `json:"item" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	Cost     int    `json:"cost,omitempty"`     // Cost in yen
	Category string `json:"category,omitempty"` // "meat", "vegetable", "seasoning", etc.
}

// DailyRecipe represents a recipe assignment for a specific day
type DailyRecipe struct {
	RecipeID int    `json:"recipe_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Day      string `json:"day,omitempty"` // monday, tuesday, etc.
}

// MealPlan represents a weekly meal plan
type MealPlan struct {
	ID        int          `json:"id" db:"id"`
	WeekData  MealPlanData `json:"week_data" db:"week_data"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// CreateMealPlanRequest represents a meal plan creation request
type CreateMealPlanRequest struct {
	StartDate   string              `json:"start_date"`
	Preferences MealPlanPreferences `json:"preferences"`
}

// MealPlanPreferences represents user preferences for meal planning
type MealPlanPreferences struct {
	MaxCookingTime      int      `json:"max_cooking_time"`
	ExcludeIngredients  []string `json:"exclude_ingredients"`
	PreferredTags       []string `json:"preferred_tags"`
	BudgetPerWeek       int      `json:"budget_per_week"`
	HouseholdSize       int      `json:"household_size"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
}

// SearchCriteria represents recipe search criteria
type SearchCriteria struct {
	Query            string  `json:"query" form:"query"` // General search query (title, ingredient)
	Tag              string  `json:"tag" form:"tag"`
	Ingredient       string  `json:"ingredient" form:"ingredient"`
	MaxCookingTime   int     `json:"max_cooking_time" form:"max_cooking_time"`
	MinLazinessScore float64 `json:"min_laziness_score" form:"min_laziness_score"`
	Season           string  `json:"season" form:"season"`
	Limit            int     `json:"limit" form:"limit"`
	Offset           int     `json:"offset" form:"offset"`
	Page             int     `json:"page" form:"page"` // Page number (alternative to offset)
}

// MealPlanData holds the JSON-stored meal plan information
type MealPlanData struct {
	StartDate         string                 `json:"start_date" binding:"required"`
	ShoppingList      []ShoppingItem         `json:"shopping_list" binding:"required"`
	DailyRecipes      map[string]DailyRecipe `json:"daily_recipes" binding:"required"`
	TotalCostEstimate int                    `json:"total_cost_estimate"`
	WeekTheme         string                 `json:"week_theme,omitempty"`
	IngredientReuse   map[string][]string    `json:"ingredient_reuse,omitempty"` // ingredient -> days used
	NutritionSummary  *WeekNutritionSummary  `json:"nutrition_summary,omitempty"`
}

// WeekNutritionSummary holds weekly nutrition totals
type WeekNutritionSummary struct {
	TotalCalories     int     `json:"total_calories"`
	AvgCaloriesPerDay int     `json:"avg_calories_per_day"`
	TotalProtein      int     `json:"total_protein"`
	BalanceScore      float64 `json:"balance_score"` // 1-10, how balanced the week is
}

// Validate validates the meal plan data
func (m *MealPlanData) Validate() error {
	if m.StartDate == "" {
		return ErrInvalidStartDate
	}
	if len(m.ShoppingList) == 0 {
		return ErrEmptyShoppingList
	}
	if len(m.DailyRecipes) == 0 {
		return ErrNoDailyRecipes
	}

	// Check that we have recipes for the expected days
	expectedDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, day := range expectedDays {
		if _, exists := m.DailyRecipes[day]; !exists {
			// Allow partial week plans, but at least 3 days
			if len(m.DailyRecipes) < 3 {
				return ErrInsufficientRecipes
			}
		}
	}

	return nil
}

// CalculateTotalCost calculates the total estimated cost
func (m *MealPlanData) CalculateTotalCost() int {
	total := 0
	for _, item := range m.ShoppingList {
		total += item.Cost
	}
	return total
}

// GetIngredientUsage analyzes which ingredients are used on which days
func (m *MealPlanData) GetIngredientUsage() map[string][]string {
	usage := make(map[string][]string)

	for day := range m.DailyRecipes {
		// This would require recipe data to determine ingredients
		// For now, create a placeholder structure
		if m.IngredientReuse != nil {
			for ingredient, days := range m.IngredientReuse {
				for _, usageDay := range days {
					if usageDay == day {
						usage[ingredient] = append(usage[ingredient], day)
					}
				}
			}
		}
	}

	return usage
}

// OptimizeShoppingList groups and optimizes the shopping list
func (m *MealPlanData) OptimizeShoppingList() {
	// Group items by category
	categoryMap := make(map[string][]ShoppingItem)

	for _, item := range m.ShoppingList {
		if item.Category == "" {
			item.Category = categorizeIngredient(item.Item)
		}
		categoryMap[item.Category] = append(categoryMap[item.Category], item)
	}

	// Rebuild shopping list with optimized order
	optimizedList := []ShoppingItem{}

	// Order: vegetables -> meat -> dairy -> seasonings -> others
	categoryOrder := []string{"vegetable", "meat", "dairy", "seasoning", "grain", "others"}

	for _, category := range categoryOrder {
		if items, exists := categoryMap[category]; exists {
			optimizedList = append(optimizedList, items...)
		}
	}

	m.ShoppingList = optimizedList
}

// categorizeIngredient attempts to categorize an ingredient
func categorizeIngredient(ingredient string) string {
	// Simple categorization logic - could be expanded with a proper dictionary
	meatKeywords := []string{"豚", "鶏", "牛", "肉", "ひき肉"}
	vegetableKeywords := []string{"キャベツ", "もやし", "玉ねぎ", "ねぎ", "にんじん", "トマト"}
	seasoningKeywords := []string{"醤油", "味噌", "塩", "胡椒", "油", "みりん", "酒"}
	dairyKeywords := []string{"卵", "牛乳", "チーズ", "バター"}
	grainKeywords := []string{"米", "パン", "麺", "うどん", "そば"}

	for _, keyword := range meatKeywords {
		if contains(ingredient, keyword) {
			return "meat"
		}
	}
	for _, keyword := range vegetableKeywords {
		if contains(ingredient, keyword) {
			return "vegetable"
		}
	}
	for _, keyword := range seasoningKeywords {
		if contains(ingredient, keyword) {
			return "seasoning"
		}
	}
	for _, keyword := range dairyKeywords {
		if contains(ingredient, keyword) {
			return "dairy"
		}
	}
	for _, keyword := range grainKeywords {
		if contains(ingredient, keyword) {
			return "grain"
		}
	}

	return "others"
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ToJSON converts MealPlanData to JSON bytes
func (m *MealPlanData) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON parses JSON bytes into MealPlanData
func (m *MealPlanData) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// GetDaysCount returns the number of days in the meal plan
func (m *MealPlanData) GetDaysCount() int {
	return len(m.DailyRecipes)
}

// GetRecipeIDs returns all recipe IDs used in the meal plan
func (m *MealPlanData) GetRecipeIDs() []int {
	ids := make([]int, 0, len(m.DailyRecipes))
	for _, recipe := range m.DailyRecipes {
		ids = append(ids, recipe.RecipeID)
	}
	return ids
}
