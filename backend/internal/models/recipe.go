package models

import (
	"encoding/json"
	"strconv"
	"time"
)

// FlexibleInt is a type that can be unmarshaled from both string and int
type FlexibleInt int

// UnmarshalJSON implements json.Unmarshaler for FlexibleInt
func (fi *FlexibleInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int first
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*fi = FlexibleInt(intVal)
		return nil
	}

	// Try to unmarshal as string and convert to int
	var strVal string
	if err := json.Unmarshal(data, &strVal); err != nil {
		return err
	}

	// Convert string to int
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return err
	}

	*fi = FlexibleInt(intVal)
	return nil
}

// Int returns the int value
func (fi FlexibleInt) Int() int {
	return int(fi)
}

// Ingredient represents a recipe ingredient
type Ingredient struct {
	Name   string  `json:"name" binding:"required"`
	Amount string  `json:"amount" binding:"required"`
	Notes  *string `json:"notes"` // Optional preparation notes
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
	ServingSize   FlexibleInt    `json:"serving_size" binding:"min=1"`
	Difficulty    string         `json:"difficulty,omitempty" binding:"omitempty,oneof=easy medium hard"`
	TotalCost     int            `json:"total_cost,omitempty"` // Cost in yen
}

// RecipeSchema defines the JSON Schema for Structured Outputs
type RecipeSchema struct {
	Type                 string                    `json:"type"`
	Properties           map[string]SchemaProperty `json:"properties"`
	Required             []string                  `json:"required"`
	AdditionalProperties bool                      `json:"additionalProperties"`
}

type SchemaProperty struct {
	Type                 interface{}               `json:"type,omitempty"` // Can be string or []string
	Format               string                    `json:"format,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Minimum              *float64                  `json:"minimum,omitempty"`
	Maximum              *float64                  `json:"maximum,omitempty"`
	MinLength            *int                      `json:"minLength,omitempty"`
	MaxLength            *int                      `json:"maxLength,omitempty"`
	Items                *SchemaProperty           `json:"items,omitempty"`
	Properties           map[string]SchemaProperty `json:"properties,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	Enum                 []string                  `json:"enum,omitempty"`
	MinItems             *int                      `json:"minItems,omitempty"`
	MaxItems             *int                      `json:"maxItems,omitempty"`
	AdditionalProperties *bool                     `json:"additionalProperties,omitempty"`
}

// GetRecipeJSONSchema returns the JSON Schema for recipe generation with food safety validation
func GetRecipeJSONSchema() RecipeSchema {
	minOne := 1
	maxThree := 3
	maxFifteen := 15.0
	minOneDot := 1.0
	maxTenDot := 10.0
	maxHundred := 100
	falseVal := false // For additionalProperties

	return RecipeSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"title": {
				Type:        "string",
				Description: "Recipe title (clear and appetizing)",
				MinLength:   &minOne,
				MaxLength:   &maxHundred,
			},
			"cooking_time": {
				Type:        "integer",
				Description: "Total cooking time in minutes",
				Minimum:     &minOneDot,
				Maximum:     &maxFifteen,
			},
			"ingredients": {
				Type:        "array",
				Description: "List of ingredients with quantities",
				MinItems:    &minOne,
				Items: &SchemaProperty{
					Type: "object",
					Properties: map[string]SchemaProperty{
						"name":   {Type: "string", Description: "Ingredient name"},
						"amount": {Type: "string", Description: "Amount (e.g., '2 cups', '300g')"},
						"notes":  {Type: []string{"string", "null"}, Description: "Optional preparation notes"},
					},
					Required:             []string{"name", "amount", "notes"},
					AdditionalProperties: &falseVal, // Required for Structured Outputs
				},
			},
			"steps": {
				Type:        "array",
				Description: "Cooking steps (maximum 3 for lazy cooking)",
				MinItems:    &minOne,
				MaxItems:    &maxThree,
				Items: &SchemaProperty{
					Type:        "string",
					Description: "Clear, concise cooking step",
				},
			},
			"tags": {
				Type:        []string{"array", "null"},
				Description: "Recipe tags",
				Items: &SchemaProperty{
					Type: "string",
					Enum: []string{"簡単", "10分以内", "15分以内", "ずぼら", "一品", "和食", "洋食", "中華", "主食", "副菜", "汁物", "デザート", "丼・ワンプレート", "常備菜・作り置き", "おやつ・甘味", "その他"},
				},
			},
			"season": {
				Type:        "string",
				Description: "Applicable season",
				Enum:        []string{"spring", "summer", "fall", "winter", "all"},
			},
			"laziness_score": {
				Type:        "number",
				Description: "Laziness score (10 = easiest)",
				Minimum:     &minOneDot,
				Maximum:     &maxTenDot,
			},
			"serving_size": {
				Type:        []string{"integer", "null"},
				Description: "Number of servings",
				Minimum:     &minOneDot,
			},
			"difficulty": {
				Type:        []string{"string", "null"},
				Description: "Difficulty level",
				Enum:        []string{"easy", "medium", "hard"},
			},
			"safety_compliance": {
				Type:        "object",
				Description: "Food safety compliance information",
				Properties: map[string]SchemaProperty{
					"safe_temp_check": {
						Type:        "boolean",
						Description: "Whether safe cooking temperatures are specified",
					},
					"temp_instructions": {
						Type:        "array",
						Description: "Temperature-specific instructions",
						Items: &SchemaProperty{
							Type: "object",
							Properties: map[string]SchemaProperty{
								"ingredient":    {Type: "string"},
								"target_temp_f": {Type: "number"},
								"instruction":   {Type: "string"},
							},
							Required:             []string{"ingredient", "target_temp_f", "instruction"},
							AdditionalProperties: &falseVal, // Required for nested objects
						},
					},
					"allergen_warnings": {
						Type:        "array",
						Description: "Allergen warnings",
						Items:       &SchemaProperty{Type: "string"},
					},
				},
				Required:             []string{"safe_temp_check", "temp_instructions", "allergen_warnings"},
				AdditionalProperties: &falseVal, // Required for nested objects
			},
		},
		Required:             []string{"title", "cooking_time", "ingredients", "steps", "tags", "season", "laziness_score", "serving_size", "difficulty", "safety_compliance"},
		AdditionalProperties: false, // Required for root object
	}
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
	if r.ServingSize.Int() <= 0 {
		r.ServingSize = FlexibleInt(1) // Default serving size
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
