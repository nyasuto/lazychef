package models

import (
	"encoding/json"
	"time"
)

// Ingredient represents a recipe ingredient
type Ingredient struct {
	Name   string `json:"name" binding:"required"`
	Amount string `json:"amount" binding:"required"`
}

// NutritionInfo holds nutritional information
type NutritionInfo struct {
	Calories int `json:"calories"`
	Protein  int `json:"protein"`
	Carbs    int `json:"carbs,omitempty"`
	Fat      int `json:"fat,omitempty"`
}

// Recipe represents a cooking recipe with laziness optimization
type Recipe struct {
	ID        int        `json:"id" db:"id"`
	Data      RecipeData `json:"data" db:"data"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// RecipeData holds the JSON-stored recipe information
type RecipeData struct {
	Title         string         `json:"title" binding:"required"`
	CookingTime   int            `json:"cooking_time" binding:"required,min=1"`
	Ingredients   []Ingredient   `json:"ingredients" binding:"required,min=1"`
	Steps         []string       `json:"steps" binding:"required,min=1"`
	Tags          []string       `json:"tags"`
	Season        string         `json:"season" binding:"required,oneof=spring summer fall winter all"`
	LazinessScore float64        `json:"laziness_score" binding:"required,min=1.0,max=10.0"`
	NutritionInfo *NutritionInfo `json:"nutrition_info,omitempty"`
	ServingSize   int            `json:"serving_size" binding:"min=1"`
	Difficulty    string         `json:"difficulty,omitempty" binding:"omitempty,oneof=easy medium hard"`
	TotalCost     int            `json:"total_cost,omitempty"` // Cost in yen
}

// Validate validates the recipe data
func (r *RecipeData) Validate() error {
	if r.Title == "" {
		return ErrInvalidTitle
	}
	if r.CookingTime <= 0 {
		return ErrInvalidCookingTime
	}
	if len(r.Ingredients) == 0 {
		return ErrNoIngredients
	}
	if len(r.Steps) == 0 {
		return ErrNoSteps
	}
	if r.LazinessScore < 1.0 || r.LazinessScore > 10.0 {
		return ErrInvalidLazinessScore
	}
	if r.ServingSize <= 0 {
		r.ServingSize = 1 // Default serving size
	}
	return nil
}

// CalculateLazinessScore automatically calculates laziness score based on recipe attributes
func (r *RecipeData) CalculateLazinessScore() float64 {
	score := 0.0

	// Cooking time factor (max 3 points)
	if r.CookingTime <= 5 {
		score += 3.0
	} else if r.CookingTime <= 10 {
		score += 2.5
	} else if r.CookingTime <= 15 {
		score += 2.0
	} else if r.CookingTime <= 30 {
		score += 1.0
	}

	// Steps factor (max 3 points)
	stepCount := len(r.Steps)
	if stepCount <= 2 {
		score += 3.0
	} else if stepCount <= 3 {
		score += 2.5
	} else if stepCount <= 4 {
		score += 2.0
	} else if stepCount <= 5 {
		score += 1.0
	}

	// Ingredients factor (max 2 points)
	ingredientCount := len(r.Ingredients)
	if ingredientCount <= 3 {
		score += 2.0
	} else if ingredientCount <= 5 {
		score += 1.5
	} else if ingredientCount <= 7 {
		score += 1.0
	}

	// Tag-based bonus (max 2 points)
	for _, tag := range r.Tags {
		switch tag {
		case "ワンパン", "ワンボウル", "レンジ":
			score += 0.7
		case "包丁不要", "火を使わない":
			score += 0.5
		case "簡単", "時短":
			score += 0.3
		}
	}

	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}

	return score
}

// ToJSON converts RecipeData to JSON bytes
func (r *RecipeData) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON parses JSON bytes into RecipeData
func (r *RecipeData) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// HasTag checks if recipe has a specific tag
func (r *RecipeData) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// IsForSeason checks if recipe is suitable for a specific season
func (r *RecipeData) IsForSeason(season string) bool {
	return r.Season == "all" || r.Season == season
}

// GetIngredientNames returns slice of ingredient names
func (r *RecipeData) GetIngredientNames() []string {
	names := make([]string, len(r.Ingredients))
	for i, ingredient := range r.Ingredients {
		names[i] = ingredient.Name
	}
	return names
}
