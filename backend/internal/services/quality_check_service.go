package services

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"lazychef/internal/config"
	"lazychef/internal/models"
)

// QualityCheckService validates recipe quality and LazyChef compliance
type QualityCheckService struct {
	config     *config.OpenAIConfig
	strictMode bool
}

// QualityCheckResult represents the result of quality validation
type QualityCheckResult struct {
	OverallScore           float64       `json:"overall_score"` // 0.0 - 1.0
	Passed                 bool          `json:"passed"`
	Violations             []string      `json:"violations"`
	Warnings               []string      `json:"warnings"`
	Scores                 QualityScores `json:"scores"`
	AutoApproved           bool          `json:"auto_approved"`
	ImprovementSuggestions []string      `json:"improvement_suggestions"`
}

// QualityScores breaks down individual quality metrics
type QualityScores struct {
	Feasibility        float64 `json:"feasibility"`         // Tools/temps/times consistency
	Readability        float64 `json:"readability"`         // Step clarity and length
	PantryFriendliness float64 `json:"pantry_friendliness"` // Common ingredient usage
	LazyChefCompliance float64 `json:"lazychef_compliance"` // Adherence to lazy cooking principles
	TimeRealism        float64 `json:"time_realism"`        // Realistic cooking time estimate
	IngredientBalance  float64 `json:"ingredient_balance"`  // Reasonable ingredient count/variety
}

// NewQualityCheckService creates a new quality check service
func NewQualityCheckService(config *config.OpenAIConfig) *QualityCheckService {
	return &QualityCheckService{
		config:     config,
		strictMode: config.FoodSafetyStrictMode,
	}
}

// ValidateRecipe performs comprehensive quality validation
func (q *QualityCheckService) ValidateRecipe(recipe *models.RecipeData) (*QualityCheckResult, error) {
	result := &QualityCheckResult{
		Passed:                 true,
		Violations:             []string{},
		Warnings:               []string{},
		ImprovementSuggestions: []string{},
		Scores:                 QualityScores{},
	}

	// Calculate individual quality scores
	result.Scores.Feasibility = q.calculateFeasibilityScore(recipe, result)
	result.Scores.Readability = q.calculateReadabilityScore(recipe, result)
	result.Scores.PantryFriendliness = q.calculatePantryFriendlinessScore(recipe, result)
	result.Scores.LazyChefCompliance = q.calculateLazyChefComplianceScore(recipe, result)
	result.Scores.TimeRealism = q.calculateTimeRealismScore(recipe, result)
	result.Scores.IngredientBalance = q.calculateIngredientBalanceScore(recipe, result)

	// Calculate overall score (weighted average)
	result.OverallScore = q.calculateOverallScore(result.Scores)

	// Check if recipe passes quality thresholds
	result.Passed = result.OverallScore >= 0.7 && len(result.Violations) == 0

	// Auto-approval logic
	result.AutoApproved = result.OverallScore >= 0.85 && len(result.Violations) == 0

	// Generate improvement suggestions
	q.generateImprovementSuggestions(recipe, result)

	return result, nil
}

// calculateFeasibilityScore checks tool/temperature/time consistency
func (q *QualityCheckService) calculateFeasibilityScore(recipe *models.RecipeData, result *QualityCheckResult) float64 {
	score := 1.0
	stepsText := strings.ToLower(strings.Join([]string(recipe.Steps), " "))

	// Check for impossible time constraints
	if recipe.CookingTime < 3 && strings.Contains(stepsText, "bake") {
		result.Violations = append(result.Violations, "Baking typically requires more than 3 minutes")
		score -= 0.3
	}

	// Check for tool consistency
	requiredTools := []string{}
	if strings.Contains(stepsText, "fry") || strings.Contains(stepsText, "sauté") {
		requiredTools = append(requiredTools, "pan/skillet")
	}
	if strings.Contains(stepsText, "bake") || strings.Contains(stepsText, "roast") {
		requiredTools = append(requiredTools, "oven")
	}
	if strings.Contains(stepsText, "boil") || strings.Contains(stepsText, "simmer") {
		requiredTools = append(requiredTools, "pot")
	}

	// Penalize if too many different tools required (not lazy)
	if len(requiredTools) > 2 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Recipe requires multiple tools: %v", requiredTools))
		score -= 0.2
	}

	// Check for temperature/time contradictions
	tempRegex := regexp.MustCompile(`(\d+)°?[fF]`)
	matches := tempRegex.FindAllStringSubmatch(stepsText, -1)
	for _, match := range matches {
		if temp, err := strconv.ParseFloat(match[1], 64); err == nil {
			if temp > 500 {
				result.Violations = append(result.Violations, fmt.Sprintf("Temperature %s°F seems too high for home cooking", match[1]))
				score -= 0.2
			}
		}
	}

	return math.Max(0, score)
}

// calculateReadabilityScore evaluates step clarity and length
func (q *QualityCheckService) calculateReadabilityScore(recipe *models.RecipeData, result *QualityCheckResult) float64 {
	score := 1.0

	// Check step count (LazyChef constraint: max 3 steps)
	if len([]string(recipe.Steps)) > 3 {
		result.Violations = append(result.Violations, fmt.Sprintf("Too many steps (%d) - LazyChef requires ≤3", len([]string(recipe.Steps))))
		score -= 0.4
	}

	// Check individual step length
	for i, step := range []string(recipe.Steps) {
		if len(step) > 200 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Step %d is quite long (%d chars) - consider simplifying", i+1, len(step)))
			score -= 0.1
		}

		// Check for clarity indicators
		if !strings.Contains(strings.ToLower(step), "until") &&
			!strings.Contains(strings.ToLower(step), "about") &&
			!regexp.MustCompile(`\d+`).MatchString(step) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Step %d could be more specific with times/measurements", i+1))
			score -= 0.1
		}
	}

	return math.Max(0, score)
}

// calculatePantryFriendlinessScore evaluates ingredient accessibility
func (q *QualityCheckService) calculatePantryFriendlinessScore(recipe *models.RecipeData, result *QualityCheckResult) float64 {
	commonIngredients := map[string]bool{
		"salt": true, "pepper": true, "oil": true, "butter": true, "onion": true,
		"garlic": true, "egg": true, "eggs": true, "flour": true, "sugar": true,
		"milk": true, "cheese": true, "chicken": true, "beef": true, "pork": true,
		"rice": true, "pasta": true, "bread": true, "tomato": true, "potato": true,
		"soy sauce": true, "vinegar": true, "lemon": true, "herbs": true, "spices": true,
	}

	exoticIngredients := []string{
		"truffle", "caviar", "foie gras", "miso paste", "tahini", "pomegranate molasses",
		"sumac", "za'atar", "harissa", "gochujang", "yuzu", "bonito flakes",
	}

	totalIngredients := len(recipe.Ingredients)
	commonCount := 0
	exoticCount := 0

	for _, ingredient := range recipe.Ingredients {
		name := strings.ToLower(ingredient.Name)

		// Check for common ingredients
		for common := range commonIngredients {
			if strings.Contains(name, common) {
				commonCount++
				break
			}
		}

		// Check for exotic ingredients
		for _, exotic := range exoticIngredients {
			if strings.Contains(name, exotic) {
				exoticCount++
				result.Warnings = append(result.Warnings, fmt.Sprintf("Ingredient '%s' may not be commonly available", ingredient.Name))
				break
			}
		}
	}

	// Calculate score based on pantry friendliness
	score := float64(commonCount) / float64(totalIngredients)

	// Penalize exotic ingredients
	if exoticCount > 0 {
		score -= float64(exoticCount) * 0.2
	}

	// Bonus for very accessible recipes
	if score > 0.8 {
		score += 0.1
	}

	return math.Max(0, math.Min(1, score))
}

// calculateLazyChefComplianceScore evaluates adherence to lazy cooking principles
func (q *QualityCheckService) calculateLazyChefComplianceScore(recipe *models.RecipeData, result *QualityCheckResult) float64 {
	score := 1.0
	stepsText := strings.ToLower(strings.Join([]string(recipe.Steps), " "))

	// Check cooking time (preferred ≤15 minutes)
	if recipe.CookingTime > 15 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Cooking time %d minutes exceeds LazyChef preference (≤15 min)", recipe.CookingTime))
		score -= 0.2
	}

	// Check for non-lazy techniques
	nonLazyTechniques := []string{
		"knead", "whip", "fold", "temper", "clarify", "julienne", "brunoise",
		"flambé", "reduce by half", "double boiler", "water bath",
	}

	for _, technique := range nonLazyTechniques {
		if strings.Contains(stepsText, technique) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Technique '%s' is not very lazy-friendly", technique))
			score -= 0.15
		}
	}

	// Bonus for lazy-friendly methods
	lazyMethods := []string{
		"microwave", "one pot", "sheet pan", "slow cooker", "no cook",
		"dump and stir", "throw together", "mix everything",
	}

	lazyBonus := 0.0
	for _, method := range lazyMethods {
		if strings.Contains(stepsText, method) {
			lazyBonus += 0.1
		}
	}

	score += lazyBonus

	// Check laziness score consistency
	if recipe.LazinessScore < 7.0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Laziness score %.1f is below LazyChef target (≥7.0)", recipe.LazinessScore))
		score -= 0.1
	}

	return math.Max(0, math.Min(1, score))
}

// calculateTimeRealismScore evaluates realistic cooking time estimates
func (q *QualityCheckService) calculateTimeRealismScore(recipe *models.RecipeData, result *QualityCheckResult) float64 {
	score := 1.0
	stepsText := strings.ToLower(strings.Join([]string(recipe.Steps), " "))

	// Extract any time mentions from steps
	timeRegex := regexp.MustCompile(`(\d+)\s*(?:min|minute|hour|hr)`)
	matches := timeRegex.FindAllStringSubmatch(stepsText, -1)

	totalStepTime := 0
	for _, match := range matches {
		if minutes, err := strconv.Atoi(match[1]); err == nil {
			if strings.Contains(match[0], "hour") || strings.Contains(match[0], "hr") {
				totalStepTime += minutes * 60
			} else {
				totalStepTime += minutes
			}
		}
	}

	// Check if declared cooking time is realistic compared to step times
	if totalStepTime > 0 {
		ratio := float64(recipe.CookingTime) / float64(totalStepTime)
		if ratio < 0.5 || ratio > 2.0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Declared cooking time %d min doesn't match step times (~%d min)",
					recipe.CookingTime, totalStepTime))
			score -= 0.3
		}
	}

	// Check for unrealistic combinations
	if recipe.CookingTime < 5 && strings.Contains(stepsText, "cook until tender") {
		result.Violations = append(result.Violations, "Cooking 'until tender' typically takes more than 5 minutes")
		score -= 0.4
	}

	return math.Max(0, score)
}

// calculateIngredientBalanceScore evaluates ingredient count and variety
func (q *QualityCheckService) calculateIngredientBalanceScore(recipe *models.RecipeData, result *QualityCheckResult) float64 {
	score := 1.0
	ingredientCount := len(recipe.Ingredients)

	// Too few ingredients
	if ingredientCount < 3 {
		result.Warnings = append(result.Warnings, "Recipe has very few ingredients - might lack flavor complexity")
		score -= 0.2
	}

	// Too many ingredients (not lazy)
	if ingredientCount > 10 {
		result.Warnings = append(result.Warnings, "Recipe has many ingredients - consider simplifying for laziness")
		score -= 0.3
	}

	// Check for ingredient variety
	categories := map[string]int{
		"protein": 0, "vegetable": 0, "starch": 0, "dairy": 0, "seasoning": 0,
	}

	for _, ingredient := range recipe.Ingredients {
		name := strings.ToLower(ingredient.Name)

		// Categorize ingredients (simplified)
		if strings.Contains(name, "chicken") || strings.Contains(name, "beef") ||
			strings.Contains(name, "pork") || strings.Contains(name, "fish") ||
			strings.Contains(name, "egg") {
			categories["protein"]++
		} else if strings.Contains(name, "onion") || strings.Contains(name, "tomato") ||
			strings.Contains(name, "pepper") || strings.Contains(name, "carrot") {
			categories["vegetable"]++
		} else if strings.Contains(name, "rice") || strings.Contains(name, "pasta") ||
			strings.Contains(name, "bread") || strings.Contains(name, "potato") {
			categories["starch"]++
		} else if strings.Contains(name, "cheese") || strings.Contains(name, "milk") ||
			strings.Contains(name, "butter") || strings.Contains(name, "cream") {
			categories["dairy"]++
		} else if strings.Contains(name, "salt") || strings.Contains(name, "pepper") ||
			strings.Contains(name, "herb") || strings.Contains(name, "spice") {
			categories["seasoning"]++
		}
	}

	// Bonus for balanced categories
	nonZeroCategories := 0
	for _, count := range categories {
		if count > 0 {
			nonZeroCategories++
		}
	}

	if nonZeroCategories >= 3 {
		score += 0.1
	}

	return math.Max(0, math.Min(1, score))
}

// calculateOverallScore computes weighted average of all scores
func (q *QualityCheckService) calculateOverallScore(scores QualityScores) float64 {
	weights := map[string]float64{
		"feasibility":         0.25,
		"readability":         0.20,
		"pantry_friendliness": 0.15,
		"lazychef_compliance": 0.20,
		"time_realism":        0.10,
		"ingredient_balance":  0.10,
	}

	weightedSum := scores.Feasibility*weights["feasibility"] +
		scores.Readability*weights["readability"] +
		scores.PantryFriendliness*weights["pantry_friendliness"] +
		scores.LazyChefCompliance*weights["lazychef_compliance"] +
		scores.TimeRealism*weights["time_realism"] +
		scores.IngredientBalance*weights["ingredient_balance"]

	return math.Max(0, math.Min(1, weightedSum))
}

// generateImprovementSuggestions provides actionable improvement suggestions
func (q *QualityCheckService) generateImprovementSuggestions(recipe *models.RecipeData, result *QualityCheckResult) {
	suggestions := []string{}

	if result.Scores.LazyChefCompliance < 0.7 {
		suggestions = append(suggestions, "Consider simplifying cooking methods or reducing active time")
	}

	if result.Scores.Readability < 0.7 {
		suggestions = append(suggestions, "Make instructions more specific with times and measurements")
	}

	if result.Scores.PantryFriendliness < 0.6 {
		suggestions = append(suggestions, "Replace exotic ingredients with more common alternatives")
	}

	if result.Scores.Feasibility < 0.7 {
		suggestions = append(suggestions, "Check temperature and timing consistency")
	}

	if len([]string(recipe.Steps)) > 3 {
		suggestions = append(suggestions, "Combine or eliminate steps to meet 3-step maximum")
	}

	if recipe.CookingTime > 15 {
		suggestions = append(suggestions, "Find ways to reduce cooking time to ≤15 minutes")
	}

	result.ImprovementSuggestions = suggestions
}
