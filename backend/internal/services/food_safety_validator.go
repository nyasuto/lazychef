package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"lazychef/internal/models"
)

// FoodSafetyValidator validates recipes for food safety compliance
type FoodSafetyValidator struct {
	usdaTemperatures  map[string]float64 // USDA safe minimum temperatures (°F)
	dangerousPatterns []string           // Dangerous cooking patterns to avoid
	requiredWarnings  map[string]string  // Required warnings for specific ingredients
	strictMode        bool               // Enable strict validation mode
}

// NewFoodSafetyValidator creates a new food safety validator
func NewFoodSafetyValidator(strictMode bool) *FoodSafetyValidator {
	return &FoodSafetyValidator{
		usdaTemperatures:  getUSDATemperatures(),
		dangerousPatterns: getDangerousPatterns(),
		requiredWarnings:  getRequiredWarnings(),
		strictMode:        strictMode,
	}
}

// SafetyCheckResult represents the result of a food safety check
type SafetyCheckResult struct {
	Passed           bool              `json:"passed"`
	Violations       []string          `json:"violations"`
	Warnings         []string          `json:"warnings"`
	RequiredTemps    []TempRequirement `json:"required_temps"`
	MissingTemps     []string          `json:"missing_temps"`
	AllergenWarnings []string          `json:"allergen_warnings"`
}

// TempRequirement represents a required temperature check
type TempRequirement struct {
	Ingredient  string  `json:"ingredient"`
	MinTempF    float64 `json:"min_temp_f"`
	Description string  `json:"description"`
}

// ValidateRecipe performs comprehensive food safety validation
func (v *FoodSafetyValidator) ValidateRecipe(recipe *models.RecipeData) (*SafetyCheckResult, error) {
	result := &SafetyCheckResult{
		Passed:           true,
		Violations:       []string{},
		Warnings:         []string{},
		RequiredTemps:    []TempRequirement{},
		MissingTemps:     []string{},
		AllergenWarnings: []string{},
	}

	// Check for dangerous patterns in steps
	v.checkDangerousPatterns(recipe, result)

	// Check temperature requirements
	v.checkTemperatureRequirements(recipe, result)

	// Check for allergen warnings
	v.checkAllergenRequirements(recipe, result)

	// Check raw ingredient handling
	v.checkRawIngredientHandling(recipe, result)

	// In strict mode, any violation fails the check
	if v.strictMode && len(result.Violations) > 0 {
		result.Passed = false
	}

	return result, nil
}

// checkDangerousPatterns checks for dangerous cooking patterns
func (v *FoodSafetyValidator) checkDangerousPatterns(recipe *models.RecipeData, result *SafetyCheckResult) {
	allText := strings.ToLower(recipe.Title + " " + strings.Join(recipe.Steps, " "))

	for _, pattern := range v.dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, allText); matched {
			result.Violations = append(result.Violations, fmt.Sprintf("Dangerous pattern detected: %s", pattern))
		}
	}
}

// checkTemperatureRequirements validates cooking temperatures
func (v *FoodSafetyValidator) checkTemperatureRequirements(recipe *models.RecipeData, result *SafetyCheckResult) {
	tempRegex := regexp.MustCompile(`(\d+)°?[fF]`)
	stepsText := strings.ToLower(strings.Join(recipe.Steps, " "))

	for _, ingredient := range recipe.Ingredients {
		ingredientName := strings.ToLower(ingredient.Name)

		// Check if this ingredient requires temperature monitoring
		if requiredTemp, exists := v.usdaTemperatures[ingredientName]; exists {
			tempReq := TempRequirement{
				Ingredient:  ingredient.Name,
				MinTempF:    requiredTemp,
				Description: fmt.Sprintf("%s must reach at least %.0f°F", ingredient.Name, requiredTemp),
			}
			result.RequiredTemps = append(result.RequiredTemps, tempReq)

			// Check if temperature is mentioned in steps
			tempMentioned := false
			matches := tempRegex.FindAllStringSubmatch(stepsText, -1)

			for _, match := range matches {
				if temp, err := strconv.ParseFloat(match[1], 64); err == nil {
					if temp >= requiredTemp {
						tempMentioned = true
						break
					}
				}
			}

			if !tempMentioned {
				result.MissingTemps = append(result.MissingTemps, ingredient.Name)
				result.Violations = append(result.Violations,
					fmt.Sprintf("Missing safe temperature instruction for %s (required: %.0f°F)",
						ingredient.Name, requiredTemp))
			}
		}
	}
}

// checkAllergenRequirements checks for required allergen warnings
func (v *FoodSafetyValidator) checkAllergenRequirements(recipe *models.RecipeData, result *SafetyCheckResult) {
	for _, ingredient := range recipe.Ingredients {
		ingredientName := strings.ToLower(ingredient.Name)

		for allergenIngredient, warning := range v.requiredWarnings {
			if strings.Contains(ingredientName, allergenIngredient) {
				result.AllergenWarnings = append(result.AllergenWarnings, warning)
			}
		}
	}
}

// checkRawIngredientHandling checks for proper raw ingredient handling
func (v *FoodSafetyValidator) checkRawIngredientHandling(recipe *models.RecipeData, result *SafetyCheckResult) {
	stepsText := strings.ToLower(strings.Join(recipe.Steps, " "))

	rawIngredients := []string{"raw chicken", "raw beef", "raw pork", "raw fish", "raw egg"}

	for _, ingredient := range recipe.Ingredients {
		ingredientName := strings.ToLower(ingredient.Name)

		for _, rawItem := range rawIngredients {
			if strings.Contains(ingredientName, strings.Replace(rawItem, "raw ", "", 1)) {
				// Check if proper handling is mentioned
				if !strings.Contains(stepsText, "wash hands") &&
					!strings.Contains(stepsText, "sanitize") &&
					!strings.Contains(stepsText, "separate") {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("Consider adding hand washing/sanitizing instructions when handling %s", ingredient.Name))
				}
			}
		}
	}
}

// getUSDATemperatures returns USDA safe minimum internal temperatures (°F)
func getUSDATemperatures() map[string]float64 {
	return map[string]float64{
		"chicken":        165, // Poultry
		"turkey":         165,
		"duck":           165,
		"beef":           145, // Beef, pork, lamb (with 3-min rest)
		"pork":           145,
		"lamb":           145,
		"ground beef":    160, // Ground meats
		"ground pork":    160,
		"ground chicken": 165,
		"ground turkey":  165,
		"fish":           145, // Fish and shellfish
		"salmon":         145,
		"tuna":           145,
		"shrimp":         145,
		"egg":            160, // Egg dishes
		"eggs":           160,
	}
}

// getDangerousPatterns returns regex patterns for dangerous cooking practices
func getDangerousPatterns() []string {
	return []string{
		`raw\s+flour`,              // Raw flour consumption
		`no.cook.*chicken`,         // No-cook chicken dishes
		`room\s+temperature.*meat`, // Leaving meat at room temperature
		`thaw.*counter`,            // Thawing on counter
		`rinse.*raw\s+chicken`,     // Rinsing raw chicken (spreads bacteria)
		`undercooked.*egg`,         // Undercooked eggs
		`raw.*cookie\s+dough`,      // Raw cookie dough
		`marinade.*reuse`,          // Reusing marinades
	}
}

// getRequiredWarnings returns required allergen warnings
func getRequiredWarnings() map[string]string {
	return map[string]string{
		"peanut":    "Contains peanuts - major allergen",
		"tree nut":  "Contains tree nuts - major allergen",
		"almond":    "Contains tree nuts - major allergen",
		"walnut":    "Contains tree nuts - major allergen",
		"milk":      "Contains milk - major allergen",
		"cheese":    "Contains milk - major allergen",
		"butter":    "Contains milk - major allergen",
		"egg":       "Contains eggs - major allergen",
		"soy":       "Contains soy - major allergen",
		"wheat":     "Contains wheat/gluten - major allergen",
		"flour":     "Contains wheat/gluten - major allergen",
		"fish":      "Contains fish - major allergen",
		"shellfish": "Contains shellfish - major allergen",
		"shrimp":    "Contains shellfish - major allergen",
	}
}

// IsTemperatureSafe checks if a given temperature meets safety requirements for an ingredient
func (v *FoodSafetyValidator) IsTemperatureSafe(ingredient string, tempF float64) bool {
	if requiredTemp, exists := v.usdaTemperatures[strings.ToLower(ingredient)]; exists {
		return tempF >= requiredTemp
	}
	return true // No specific requirement
}

// GetRequiredTemperature returns the USDA required temperature for an ingredient
func (v *FoodSafetyValidator) GetRequiredTemperature(ingredient string) (float64, bool) {
	temp, exists := v.usdaTemperatures[strings.ToLower(ingredient)]
	return temp, exists
}
