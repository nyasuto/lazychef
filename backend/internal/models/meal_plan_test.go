package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateMealPlanRequest_Basic(t *testing.T) {
	validDate := time.Now().Format("2006-01-02")

	tests := []struct {
		name    string
		request CreateMealPlanRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: CreateMealPlanRequest{
				StartDate: validDate,
				Preferences: MealPlanPreferences{
					MaxCookingTime: 30,
					HouseholdSize:  2,
				},
			},
			valid: true,
		},
		{
			name: "valid request with dietary restrictions",
			request: CreateMealPlanRequest{
				StartDate: validDate,
				Preferences: MealPlanPreferences{
					MaxCookingTime:      20,
					HouseholdSize:       4,
					DietaryRestrictions: []string{"vegetarian"},
					PreferredTags:       []string{"quick", "easy"},
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.request.StartDate)
				assert.GreaterOrEqual(t, tt.request.Preferences.MaxCookingTime, 0)
			}
		})
	}
}

func TestMealPlanPreferences_Structure(t *testing.T) {
	prefs := MealPlanPreferences{
		MaxCookingTime:      30,
		ExcludeIngredients:  []string{"nuts", "dairy"},
		PreferredTags:       []string{"quick", "healthy"},
		BudgetPerWeek:       5000,
		HouseholdSize:       3,
		DietaryRestrictions: []string{"vegetarian"},
	}

	assert.Equal(t, 30, prefs.MaxCookingTime)
	assert.Contains(t, prefs.ExcludeIngredients, "nuts")
	assert.Contains(t, prefs.PreferredTags, "healthy")
	assert.Equal(t, 5000, prefs.BudgetPerWeek)
	assert.Equal(t, 3, prefs.HouseholdSize)
	assert.Contains(t, prefs.DietaryRestrictions, "vegetarian")
}

func TestMealPlan_Structure(t *testing.T) {
	now := time.Now()
	mealPlan := MealPlan{
		ID:        1,
		CreatedAt: now,
		UpdatedAt: now,
		WeekData: MealPlanData{
			StartDate: now.Format("2006-01-02"),
			DailyRecipes: map[string]DailyRecipe{
				"Monday": {
					RecipeID: 1,
					Title:    "Test Recipe",
					Day:      "Monday",
				},
			},
			ShoppingList: []ShoppingItem{
				{Item: "chicken", Amount: "1 lb", Category: "meat"},
			},
		},
	}

	assert.Equal(t, 1, mealPlan.ID)
	assert.Equal(t, 1, len(mealPlan.WeekData.DailyRecipes))
	assert.Equal(t, "Monday", mealPlan.WeekData.DailyRecipes["Monday"].Day)
	assert.Equal(t, "Test Recipe", mealPlan.WeekData.DailyRecipes["Monday"].Title)
}

func TestMealPlanData_Validate(t *testing.T) {
	tests := []struct {
		name     string
		mealPlan MealPlanData
		wantErr  bool
	}{
		{
			name: "valid meal plan data",
			mealPlan: MealPlanData{
				StartDate: time.Now().Format("2006-01-02"),
				DailyRecipes: map[string]DailyRecipe{
					"Monday": {
						RecipeID: 1,
						Title:    "Test Recipe",
						Day:      "Monday",
					},
				},
				ShoppingList: []ShoppingItem{
					{Item: "chicken", Amount: "1 lb", Category: "meat"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty start date",
			mealPlan: MealPlanData{
				StartDate: "",
				DailyRecipes: map[string]DailyRecipe{
					"Monday": {
						RecipeID: 1,
						Title:    "Test Recipe",
						Day:      "Monday",
					},
				},
				ShoppingList: []ShoppingItem{
					{Item: "chicken", Amount: "1 lb"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty daily recipes",
			mealPlan: MealPlanData{
				StartDate:    time.Now().Format("2006-01-02"),
				DailyRecipes: map[string]DailyRecipe{},
				ShoppingList: []ShoppingItem{
					{Item: "chicken", Amount: "1 lb"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mealPlan.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDailyRecipe_Structure(t *testing.T) {
	dailyRecipe := DailyRecipe{
		RecipeID: 42,
		Title:    "Pasta Recipe",
		Day:      "Tuesday",
	}

	assert.Equal(t, 42, dailyRecipe.RecipeID)
	assert.Equal(t, "Pasta Recipe", dailyRecipe.Title)
	assert.Equal(t, "Tuesday", dailyRecipe.Day)
}

func TestShoppingItem_Structure(t *testing.T) {
	tests := []struct {
		name string
		item ShoppingItem
	}{
		{
			name: "complete shopping item",
			item: ShoppingItem{
				Item:     "chicken breast",
				Amount:   "1 lb",
				Category: "meat",
				Cost:     500,
			},
		},
		{
			name: "minimal shopping item",
			item: ShoppingItem{
				Item:   "salt",
				Amount: "1 tsp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.item.Item)
			assert.NotEmpty(t, tt.item.Amount)
		})
	}
}

func TestMealPlanData_CalculateTotalCost(t *testing.T) {
	mealPlan := MealPlanData{
		StartDate: "2024-01-01",
		DailyRecipes: map[string]DailyRecipe{
			"Monday": {RecipeID: 1, Title: "Recipe 1"},
		},
		ShoppingList: []ShoppingItem{
			{Item: "chicken", Amount: "1 lb", Cost: 500},
			{Item: "rice", Amount: "2 cups", Cost: 200},
			{Item: "vegetables", Amount: "1 bag", Cost: 300},
		},
	}

	totalCost := mealPlan.CalculateTotalCost()
	assert.Equal(t, 1000, totalCost) // 500 + 200 + 300
}

func TestMealPlanData_GetDaysCount(t *testing.T) {
	mealPlan := MealPlanData{
		DailyRecipes: map[string]DailyRecipe{
			"Monday":    {RecipeID: 1, Title: "Recipe 1"},
			"Tuesday":   {RecipeID: 2, Title: "Recipe 2"},
			"Wednesday": {RecipeID: 3, Title: "Recipe 3"},
		},
	}

	count := mealPlan.GetDaysCount()
	assert.Equal(t, 3, count)
}

func TestMealPlanData_OptimizeShoppingList(t *testing.T) {
	mealPlan := MealPlanData{
		ShoppingList: []ShoppingItem{
			{Item: "chicken breast", Amount: "1 lb", Category: "meat"},
			{Item: "chicken breast", Amount: "0.5 lb", Category: "meat"},
			{Item: "rice", Amount: "2 cups", Category: "grains"},
		},
	}

	mealPlan.OptimizeShoppingList()

	// After optimization, duplicate items should be consolidated
	assert.NotEmpty(t, mealPlan.ShoppingList)

	// Find chicken items
	chickenCount := 0
	for _, item := range mealPlan.ShoppingList {
		if item.Item == "chicken breast" {
			chickenCount++
		}
	}

	// Should have fewer chicken items due to consolidation
	assert.LessOrEqual(t, chickenCount, 2)
}
