package services

import (
	"fmt"
	"math/rand"

	"lazychef/internal/database"
	"lazychef/internal/models"
)

// AutoGenerationService handles AI-powered automatic recipe generation
type AutoGenerationService struct {
	db               *database.Database
	diversityService *DiversityService
}

// DimensionCombination represents a specific combination of dimensions
type DimensionCombination struct {
	MealType      *models.RecipeDimension `json:"meal_type,omitempty"`
	Protein       *models.RecipeDimension `json:"protein,omitempty"`
	CookingMethod *models.RecipeDimension `json:"cooking_method,omitempty"`
	Seasoning     *models.RecipeDimension `json:"seasoning,omitempty"`
	LazynessLevel *models.RecipeDimension `json:"laziness_level,omitempty"`
}

// CoverageAnalysis represents analysis of recipe coverage
type CoverageAnalysis struct {
	TotalCombinations     int                          `json:"total_combinations"`
	CoveredCombinations   int                          `json:"covered_combinations"`
	CoveragePercentage    float64                      `json:"coverage_percentage"`
	UncoveredCombinations []DimensionCombination       `json:"uncovered_combinations"`
	PriorityTargets       []DimensionCombination       `json:"priority_targets"`
	DimensionTypeAnalysis map[string]DimensionAnalysis `json:"dimension_type_analysis"`
}

// DimensionAnalysis represents coverage analysis for a specific dimension type
type DimensionAnalysis struct {
	DimensionType string                   `json:"dimension_type"`
	TotalValues   int                      `json:"total_values"`
	CoveredValues int                      `json:"covered_values"`
	Coverage      float64                  `json:"coverage"`
	MissingValues []models.RecipeDimension `json:"missing_values"`
}

// GenerationStrategy represents different approaches for recipe generation
type GenerationStrategy string

const (
	StrategyDiversityGapFill GenerationStrategy = "diversity_gap_fill"
	StrategyRandom           GenerationStrategy = "random"
	StrategyWeighted         GenerationStrategy = "weighted"
)

// AutoGenerationRequest represents a request for automatic recipe generation
type AutoGenerationRequest struct {
	Count             int                `json:"count" binding:"required,min=1,max=50"`
	Strategy          GenerationStrategy `json:"strategy"`
	MaxCookingTime    int                `json:"max_cooking_time,omitempty"`
	TargetLazinessMin float64            `json:"target_laziness_min,omitempty"`
	ForcedDimensions  map[string]string  `json:"forced_dimensions,omitempty"`
}

// NewAutoGenerationService creates a new auto generation service
func NewAutoGenerationService(db *database.Database, diversityService *DiversityService) *AutoGenerationService {
	return &AutoGenerationService{
		db:               db,
		diversityService: diversityService,
	}
}

// AnalyzeCoverage analyzes the current recipe coverage across all dimensions
func (s *AutoGenerationService) AnalyzeCoverage() (*CoverageAnalysis, error) {
	// Get all dimension types and their values
	dimensionTypes := []string{"meal_type", "protein", "cooking_method", "seasoning", "laziness_level"}

	analysis := &CoverageAnalysis{
		DimensionTypeAnalysis: make(map[string]DimensionAnalysis),
	}

	// For each dimension type, analyze coverage
	for _, dimType := range dimensionTypes {
		dimAnalysis, err := s.analyzeDimensionTypeCoverage(dimType)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze %s coverage: %w", dimType, err)
		}
		analysis.DimensionTypeAnalysis[dimType] = *dimAnalysis
	}

	// Generate priority targets (combinations that should be generated first)
	priorityTargets, err := s.generatePriorityTargets(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate priority targets: %w", err)
	}
	analysis.PriorityTargets = priorityTargets

	// Calculate overall coverage
	totalRecipes, err := s.getTotalRecipeCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get total recipe count: %w", err)
	}

	// Estimate total possible meaningful combinations (not all 26^5, but reasonable ones)
	analysis.TotalCombinations = s.estimateMeaningfulCombinations()
	analysis.CoveredCombinations = totalRecipes
	if analysis.TotalCombinations > 0 {
		analysis.CoveragePercentage = float64(analysis.CoveredCombinations) / float64(analysis.TotalCombinations) * 100
	}

	return analysis, nil
}

// analyzeDimensionTypeCoverage analyzes coverage for a specific dimension type
func (s *AutoGenerationService) analyzeDimensionTypeCoverage(dimensionType string) (*DimensionAnalysis, error) {
	// Get all dimensions of this type
	query := `SELECT id, dimension_value FROM recipe_dimensions WHERE dimension_type = ? AND is_active = 1`
	rows, err := s.db.Query(query, dimensionType)
	if err != nil {
		return nil, fmt.Errorf("failed to query dimensions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log the error or handle appropriately in production
		}
	}()

	var allDimensions []models.RecipeDimension
	for rows.Next() {
		var dim models.RecipeDimension
		if err := rows.Scan(&dim.ID, &dim.DimensionValue); err != nil {
			return nil, fmt.Errorf("failed to scan dimension: %w", err)
		}
		dim.DimensionType = dimensionType
		allDimensions = append(allDimensions, dim)
	}

	// For now, assume basic coverage (this would be enhanced with actual recipe-dimension mappings)
	analysis := &DimensionAnalysis{
		DimensionType: dimensionType,
		TotalValues:   len(allDimensions),
		CoveredValues: int(float64(len(allDimensions)) * 0.3), // Placeholder: assume 30% coverage
		MissingValues: []models.RecipeDimension{},
	}

	if analysis.TotalValues > 0 {
		analysis.Coverage = float64(analysis.CoveredValues) / float64(analysis.TotalValues) * 100
	}

	// Add some missing values as examples (in real implementation, this would be based on actual mappings)
	if len(allDimensions) > 0 {
		analysis.MissingValues = allDimensions[analysis.CoveredValues:]
	}

	return analysis, nil
}

// generatePriorityTargets generates priority dimension combinations for recipe generation
func (s *AutoGenerationService) generatePriorityTargets(count int) ([]DimensionCombination, error) {
	var targets []DimensionCombination

	// Get dimensions by type
	dimensionsByType, err := s.getDimensionsByType()
	if err != nil {
		return nil, fmt.Errorf("failed to get dimensions by type: %w", err)
	}

	// Generate strategic combinations
	for i := 0; i < count && i < 50; i++ {
		combination := DimensionCombination{}

		// Pick high-weight dimensions with some randomness
		if mealTypes, ok := dimensionsByType["meal_type"]; ok && len(mealTypes) > 0 {
			combination.MealType = &mealTypes[rand.Intn(len(mealTypes))]
		}
		if proteins, ok := dimensionsByType["protein"]; ok && len(proteins) > 0 {
			combination.Protein = &proteins[rand.Intn(len(proteins))]
		}
		if methods, ok := dimensionsByType["cooking_method"]; ok && len(methods) > 0 {
			combination.CookingMethod = &methods[rand.Intn(len(methods))]
		}
		if seasonings, ok := dimensionsByType["seasoning"]; ok && len(seasonings) > 0 {
			combination.Seasoning = &seasonings[rand.Intn(len(seasonings))]
		}
		if laziness, ok := dimensionsByType["laziness_level"]; ok && len(laziness) > 0 {
			combination.LazynessLevel = &laziness[rand.Intn(len(laziness))]
		}

		targets = append(targets, combination)
	}

	return targets, nil
}

// getDimensionsByType retrieves dimensions organized by type
func (s *AutoGenerationService) getDimensionsByType() (map[string][]models.RecipeDimension, error) {
	query := `SELECT id, dimension_type, dimension_value, weight FROM recipe_dimensions WHERE is_active = 1 ORDER BY weight DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query dimensions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log the error or handle appropriately in production
		}
	}()

	dimensionsByType := make(map[string][]models.RecipeDimension)
	for rows.Next() {
		var dim models.RecipeDimension
		if err := rows.Scan(&dim.ID, &dim.DimensionType, &dim.DimensionValue, &dim.Weight); err != nil {
			return nil, fmt.Errorf("failed to scan dimension: %w", err)
		}

		dimensionsByType[dim.DimensionType] = append(dimensionsByType[dim.DimensionType], dim)
	}

	return dimensionsByType, nil
}

// getTotalRecipeCount gets the total number of recipes
func (s *AutoGenerationService) getTotalRecipeCount() (int, error) {
	var count int
	row := s.db.QueryRow("SELECT COUNT(*) FROM recipes")
	err := row.Scan(&count)
	return count, err
}

// estimateMeaningfulCombinations estimates the number of meaningful dimension combinations
func (s *AutoGenerationService) estimateMeaningfulCombinations() int {
	// Not all 26^5 combinations make sense
	// Estimate based on practical cooking combinations:
	// - 7 meal types × 7 proteins × 5 cooking methods × 5 seasonings × 3 laziness = 3,675
	// But many combinations don't make sense, so use a more conservative estimate
	return 500 // Reasonable target for a comprehensive recipe collection
}

// DimensionCombinationToIngredients converts a dimension combination to suitable ingredients
func (s *AutoGenerationService) DimensionCombinationToIngredients(combo DimensionCombination) []string {
	var ingredients []string

	// Add protein-based ingredient
	if combo.Protein != nil {
		switch combo.Protein.DimensionValue {
		case "鶏肉":
			ingredients = append(ingredients, "鶏胸肉")
		case "豚肉":
			ingredients = append(ingredients, "豚こま肉")
		case "牛肉":
			ingredients = append(ingredients, "牛切り落とし")
		case "卵":
			ingredients = append(ingredients, "卵")
		case "豆腐":
			ingredients = append(ingredients, "豆腐")
		case "ツナ缶":
			ingredients = append(ingredients, "ツナ缶")
		}
	}

	// Add vegetables based on meal type and cooking method
	if combo.MealType != nil && combo.CookingMethod != nil {
		switch combo.MealType.DimensionValue {
		case "主菜", "丼・ワンプレート":
			vegetables := []string{"玉ねぎ", "キャベツ", "人参", "ピーマン", "もやし"}
			ingredients = append(ingredients, vegetables[rand.Intn(len(vegetables))])
		case "副菜":
			vegetables := []string{"きゅうり", "トマト", "レタス", "白菜"}
			ingredients = append(ingredients, vegetables[rand.Intn(len(vegetables))])
		case "汁物":
			ingredients = append(ingredients, "わかめ", "ねぎ")
		}
	}

	// Add staple ingredients for main dishes
	if combo.MealType != nil {
		switch combo.MealType.DimensionValue {
		case "主食", "丼・ワンプレート":
			staples := []string{"ご飯", "パスタ", "うどん"}
			ingredients = append(ingredients, staples[rand.Intn(len(staples))])
		}
	}

	// Ensure at least 2 ingredients
	if len(ingredients) < 2 {
		defaultIngredients := []string{"玉ねぎ", "醤油", "ごま油", "にんにく"}
		for len(ingredients) < 2 && len(defaultIngredients) > 0 {
			idx := rand.Intn(len(defaultIngredients))
			ingredients = append(ingredients, defaultIngredients[idx])
			defaultIngredients = append(defaultIngredients[:idx], defaultIngredients[idx+1:]...)
		}
	}

	return ingredients
}

// GetGenerationParameters converts dimension combination to generation parameters
func (s *AutoGenerationService) GetGenerationParameters(combo DimensionCombination) RecipeGenerationRequest {
	req := RecipeGenerationRequest{
		Ingredients:    s.DimensionCombinationToIngredients(combo),
		Season:         "all",
		MaxCookingTime: 15, // Default
		Servings:       1,
	}

	// Adjust cooking time based on laziness level
	if combo.LazynessLevel != nil {
		switch combo.LazynessLevel.DimensionValue {
		case "1_超簡単":
			req.MaxCookingTime = 5
		case "2_簡単":
			req.MaxCookingTime = 10
		case "3_ちょい手間":
			req.MaxCookingTime = 20
		}
	}

	// Add constraints based on cooking method
	if combo.CookingMethod != nil {
		switch combo.CookingMethod.DimensionValue {
		case "電子レンジ":
			req.Constraints = append(req.Constraints, "電子レンジのみ使用")
		case "和えるだけ":
			req.Constraints = append(req.Constraints, "火を使わない", "和えるだけ")
		}
	}

	// Add preferences based on seasoning
	if combo.Seasoning != nil {
		req.Preferences = append(req.Preferences, combo.Seasoning.DimensionValue+"の味付け")
	}

	return req
}

// GeneratePriorityTargets is a public wrapper for generatePriorityTargets
func (s *AutoGenerationService) GeneratePriorityTargets(count int) ([]DimensionCombination, error) {
	return s.generatePriorityTargets(count)
}
