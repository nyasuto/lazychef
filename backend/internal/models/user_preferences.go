package models

import (
	"encoding/json"
	"time"
)

// UserPreferences represents user settings and preferences
type UserPreferences struct {
	ID          int                   `json:"id" db:"id"`
	UserID      string                `json:"user_id" db:"user_id"`
	Preferences UserPreferencesData   `json:"preferences" db:"preferences"`
	UpdatedAt   time.Time             `json:"updated_at" db:"updated_at"`
}

// UserPreferencesData holds the JSON-stored preferences
type UserPreferencesData struct {
	MaxCookingTime      int      `json:"max_cooking_time" binding:"min=1,max=120"`
	ExcludeIngredients  []string `json:"exclude_ingredients"`
	PreferredTags       []string `json:"preferred_tags"`
	BudgetPerWeek       int      `json:"budget_per_week" binding:"min=100"`
	HouseholdSize       int      `json:"household_size" binding:"min=1,max=10"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
	PreferredSeasons    []string `json:"preferred_seasons"`
	CookingSkillLevel   string   `json:"cooking_skill_level,omitempty" binding:"omitempty,oneof=beginner intermediate advanced"`
	KitchenEquipment    []string `json:"kitchen_equipment,omitempty"`
	AllergyInfo         []string `json:"allergy_info,omitempty"`
	FavoriteIngredients []string `json:"favorite_ingredients,omitempty"`
	MealPlanLength      int      `json:"meal_plan_length" binding:"min=1,max=14"` // days
}

// Validate validates the user preferences
func (u *UserPreferencesData) Validate() error {
	if u.MaxCookingTime <= 0 {
		u.MaxCookingTime = 15 // Default to 15 minutes
	}
	if u.BudgetPerWeek <= 0 {
		u.BudgetPerWeek = 3000 // Default budget
	}
	if u.HouseholdSize <= 0 {
		u.HouseholdSize = 1 // Default to single person
	}
	if u.MealPlanLength <= 0 {
		u.MealPlanLength = 7 // Default to 1 week
	}
	
	// Validate preferred seasons
	validSeasons := map[string]bool{
		"spring": true,
		"summer": true,
		"fall":   true,
		"winter": true,
		"all":    true,
	}
	
	validPreferredSeasons := []string{}
	for _, season := range u.PreferredSeasons {
		if validSeasons[season] {
			validPreferredSeasons = append(validPreferredSeasons, season)
		}
	}
	u.PreferredSeasons = validPreferredSeasons
	
	if len(u.PreferredSeasons) == 0 {
		u.PreferredSeasons = []string{"all"}
	}
	
	return nil
}

// IsIngredientExcluded checks if an ingredient should be excluded
func (u *UserPreferencesData) IsIngredientExcluded(ingredient string) bool {
	for _, excluded := range u.ExcludeIngredients {
		if excluded == ingredient {
			return true
		}
	}
	
	// Check against allergy info as well
	for _, allergy := range u.AllergyInfo {
		if contains(ingredient, allergy) {
			return true
		}
	}
	
	return false
}

// HasPreferredTag checks if a tag is in the preferred list
func (u *UserPreferencesData) HasPreferredTag(tag string) bool {
	for _, preferred := range u.PreferredTags {
		if preferred == tag {
			return true
		}
	}
	return false
}

// SupportsSeason checks if a season is preferred
func (u *UserPreferencesData) SupportsSeason(season string) bool {
	for _, preferred := range u.PreferredSeasons {
		if preferred == "all" || preferred == season {
			return true
		}
	}
	return false
}

// GetCookingTimeMultiplier returns a multiplier based on skill level
func (u *UserPreferencesData) GetCookingTimeMultiplier() float64 {
	switch u.CookingSkillLevel {
	case "beginner":
		return 1.5 // Beginners take 50% more time
	case "intermediate":
		return 1.0 // Normal time
	case "advanced":
		return 0.8 // Advanced cooks are 20% faster
	default:
		return 1.2 // Default assumes slight beginner tendency
	}
}

// HasKitchenEquipment checks if specific equipment is available
func (u *UserPreferencesData) HasKitchenEquipment(equipment string) bool {
	for _, item := range u.KitchenEquipment {
		if item == equipment {
			return true
		}
	}
	return false
}

// GetDefaultPreferences returns default user preferences
func GetDefaultPreferences() UserPreferencesData {
	return UserPreferencesData{
		MaxCookingTime:      15,
		ExcludeIngredients:  []string{},
		PreferredTags:       []string{"簡単", "10分以内", "ずぼら"},
		BudgetPerWeek:       3000,
		HouseholdSize:       1,
		DietaryRestrictions: []string{},
		PreferredSeasons:    []string{"all"},
		CookingSkillLevel:   "beginner",
		KitchenEquipment:    []string{"コンロ", "電子レンジ", "冷蔵庫"},
		AllergyInfo:         []string{},
		FavoriteIngredients: []string{},
		MealPlanLength:      7,
	}
}

// ToJSON converts UserPreferencesData to JSON bytes
func (u *UserPreferencesData) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

// FromJSON parses JSON bytes into UserPreferencesData
func (u *UserPreferencesData) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

// UpdateFromRequest updates preferences from a request payload
func (u *UserPreferencesData) UpdateFromRequest(update UserPreferencesData) {
	if update.MaxCookingTime > 0 {
		u.MaxCookingTime = update.MaxCookingTime
	}
	if len(update.ExcludeIngredients) > 0 {
		u.ExcludeIngredients = update.ExcludeIngredients
	}
	if len(update.PreferredTags) > 0 {
		u.PreferredTags = update.PreferredTags
	}
	if update.BudgetPerWeek > 0 {
		u.BudgetPerWeek = update.BudgetPerWeek
	}
	if update.HouseholdSize > 0 {
		u.HouseholdSize = update.HouseholdSize
	}
	if len(update.DietaryRestrictions) >= 0 { // Allow empty slice to clear restrictions
		u.DietaryRestrictions = update.DietaryRestrictions
	}
	if len(update.PreferredSeasons) > 0 {
		u.PreferredSeasons = update.PreferredSeasons
	}
	if update.CookingSkillLevel != "" {
		u.CookingSkillLevel = update.CookingSkillLevel
	}
	if len(update.KitchenEquipment) >= 0 {
		u.KitchenEquipment = update.KitchenEquipment
	}
	if len(update.AllergyInfo) >= 0 {
		u.AllergyInfo = update.AllergyInfo
	}
	if len(update.FavoriteIngredients) >= 0 {
		u.FavoriteIngredients = update.FavoriteIngredients
	}
	if update.MealPlanLength > 0 {
		u.MealPlanLength = update.MealPlanLength
	}
}