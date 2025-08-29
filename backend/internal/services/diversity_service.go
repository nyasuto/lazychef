package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"lazychef/internal/database"
	"lazychef/internal/models"
)

// DiversityService handles recipe diversity and coverage analysis
type DiversityService struct {
	db               *database.Database
	diversityRepo    *DiversityRepository
	recipeRepo       *RecipeRepository
	generatorService *RecipeGeneratorService
}

// NewDiversityService creates a new diversity service
func NewDiversityService(db *database.Database, generator *RecipeGeneratorService) *DiversityService {
	return &DiversityService{
		db:               db,
		diversityRepo:    NewDiversityRepository(db),
		recipeRepo:       NewRecipeRepository(db),
		generatorService: generator,
	}
}

// AnalyzeCoverage analyzes current recipe coverage across dimensions
func (s *DiversityService) AnalyzeCoverage() (*models.CoverageAnalysis, error) {
	// Get all coverage data
	coverages, err := s.diversityRepo.GetDimensionCoverage()
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage data: %w", err)
	}

	// Initialize coverage analysis
	analysis := &models.CoverageAnalysis{
		TotalCombinations:   len(coverages),
		CoveredCombinations: 0,
		LowCoverageCombos:   make([]models.CoverageSummary, 0),
		DimensionStats:      make(map[string]models.DimensionStat),
	}

	// Track dimension statistics
	dimensionCounts := make(map[string]map[string]int)
	dimensionTotals := make(map[string]int)

	for _, coverage := range coverages {
		if coverage.CurrentCount > 0 {
			analysis.CoveredCombinations++
		}

		// Parse combo for dimension analysis
		var combo models.DimensionCombo
		if err := combo.FromJSON(coverage.DimensionCombo); err != nil {
			continue
		}

		// Update dimension statistics
		s.updateDimensionStats(dimensionCounts, dimensionTotals, combo, coverage.CurrentCount > 0)

		// Identify low coverage combinations (less than target)
		if coverage.CurrentCount < coverage.TargetCount {
			gap := coverage.TargetCount - coverage.CurrentCount
			summary := models.CoverageSummary{
				Combo:        combo,
				CurrentCount: coverage.CurrentCount,
				TargetCount:  coverage.TargetCount,
				Priority:     coverage.PriorityScore,
				Gap:          gap,
			}
			analysis.LowCoverageCombos = append(analysis.LowCoverageCombos, summary)
		}
	}

	// Calculate coverage rate
	if analysis.TotalCombinations > 0 {
		analysis.CoverageRate = float64(analysis.CoveredCombinations) / float64(analysis.TotalCombinations)
	}

	// Calculate dimension statistics
	for dimType, counts := range dimensionCounts {
		total := dimensionTotals[dimType]
		covered := 0
		totalCoverage := 0

		for _, count := range counts {
			if count > 0 {
				covered++
			}
			totalCoverage += count
		}

		avgCoverage := 0.0
		if len(counts) > 0 {
			avgCoverage = float64(totalCoverage) / float64(len(counts))
		}

		analysis.DimensionStats[dimType] = models.DimensionStat{
			DimensionType: dimType,
			TotalValues:   total,
			CoveredValues: covered,
			AvgCoverage:   avgCoverage,
		}
	}

	return analysis, nil
}

// updateDimensionStats updates dimension statistics for coverage analysis
func (s *DiversityService) updateDimensionStats(counts map[string]map[string]int, totals map[string]int, combo models.DimensionCombo, hasCoverage bool) {
	dimensions := map[string]string{
		"meal_type":      combo.MealType,
		"staple":         combo.Staple,
		"protein":        combo.Protein,
		"cooking_method": combo.CookingMethod,
		"seasoning":      combo.Seasoning,
		"laziness_level": combo.LazynessLevel,
	}

	for dimType, dimValue := range dimensions {
		if counts[dimType] == nil {
			counts[dimType] = make(map[string]int)
		}

		counts[dimType][dimValue]++
		totals[dimType]++
	}
}

// GenerateDiverseRecipes generates recipes using diversity-focused strategies
func (s *DiversityService) GenerateDiverseRecipes(req models.DiverseGenerationRequest) (*models.DiverseGenerationResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get generation profile
	profile, err := s.diversityRepo.GetGenerationProfile(req.ProfileName)
	if err != nil {
		return nil, err
	}

	// Parse profile configuration
	var config models.GenerationConfig
	if err := json.Unmarshal(profile.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse profile config: %w", err)
	}

	// Apply request overrides
	s.applyRequestOverrides(&config, req)

	// Generate target combinations based on strategy
	targetCombos, err := s.selectTargetCombinations(config)
	if err != nil {
		return nil, fmt.Errorf("failed to select target combinations: %w", err)
	}

	// Limit to requested batch size
	if len(targetCombos) > req.BatchSize {
		targetCombos = targetCombos[:req.BatchSize]
	}

	// Generate recipes for target combinations
	generatedRecipes := make([]models.RecipeData, 0)
	diversityScore := 0.0

	for i, combo := range targetCombos {
		// Create generation prompt based on combination
		prompt := s.createPromptFromCombo(combo, config)

		// Generate recipe (simplified - in real implementation would use batch API)
		recipeData, err := s.generateSingleRecipe(prompt, config)
		if err != nil {
			log.Printf("Warning: failed to generate recipe for combo %v: %v", combo, err)
			continue
		}

		generatedRecipes = append(generatedRecipes, *recipeData)

		// Update coverage
		comboJSON, _ := combo.ToJSON()
		if err := s.diversityRepo.UpsertDimensionCoverage(comboJSON, 1); err != nil {
			log.Printf("Warning: failed to update coverage for combo %s: %v", comboJSON, err)
		}

		// Calculate diversity score (simplified)
		diversityScore += float64(i+1) / float64(len(targetCombos))
	}

	// Calculate coverage impact
	impact := models.CoverageImpact{
		NewCombinations:      len(generatedRecipes), // Simplified
		ImprovedCombinations: len(generatedRecipes),
		TotalCombinations:    len(targetCombos),
	}

	response := &models.DiverseGenerationResponse{
		JobID:          fmt.Sprintf("diverse_%d", time.Now().Unix()),
		ProfileUsed:    req.ProfileName,
		Strategy:       config.Strategy,
		RequestedCount: req.BatchSize,
		GeneratedCount: len(generatedRecipes),
		DiversityScore: diversityScore / float64(len(targetCombos)),
		CoverageImpact: impact,
		EstimatedCost:  float64(len(generatedRecipes)) * 0.01, // $0.01 per recipe estimate
		Recipes:        generatedRecipes,
	}

	return response, nil
}

// applyRequestOverrides applies request-specific overrides to the configuration
func (s *DiversityService) applyRequestOverrides(config *models.GenerationConfig, req models.DiverseGenerationRequest) {
	if req.Strategy != "" {
		config.Strategy = req.Strategy
	}
	if req.MaxSimilarity != nil {
		config.MaxSimilarity = *req.MaxSimilarity
	}
	if req.QualityThreshold != nil {
		config.QualityThreshold = *req.QualityThreshold
	}
	if len(req.CustomWeights) > 0 {
		if config.DimensionWeights == nil {
			config.DimensionWeights = make(map[string]float64)
		}
		for key, weight := range req.CustomWeights {
			config.DimensionWeights[key] = weight
		}
	}
}

// selectTargetCombinations selects target combinations based on strategy
func (s *DiversityService) selectTargetCombinations(config models.GenerationConfig) ([]models.DimensionCombo, error) {
	switch config.Strategy {
	case "coverage_first":
		return s.selectByCoverage(config.BatchSize)
	case "priority_first":
		return s.selectByPriority(config.BatchSize)
	case "random_sample":
		return s.selectRandomly(config.BatchSize)
	default:
		return s.selectByCoverage(config.BatchSize) // Default to coverage-first
	}
}

// selectByCoverage selects combinations with lowest coverage first
func (s *DiversityService) selectByCoverage(batchSize int) ([]models.DimensionCombo, error) {
	lowCoverage, err := s.diversityRepo.GetLowCoverageCombinations(batchSize)
	if err != nil {
		return nil, err
	}

	combos := make([]models.DimensionCombo, 0, len(lowCoverage))
	for _, coverage := range lowCoverage {
		var combo models.DimensionCombo
		if err := combo.FromJSON(coverage.DimensionCombo); err != nil {
			continue
		}
		combos = append(combos, combo)
	}

	return combos, nil
}

// selectByPriority selects combinations by priority score
func (s *DiversityService) selectByPriority(batchSize int) ([]models.DimensionCombo, error) {
	// Similar to coverage, but considers priority weighting
	return s.selectByCoverage(batchSize) // Simplified implementation
}

// selectRandomly selects combinations randomly
func (s *DiversityService) selectRandomly(batchSize int) ([]models.DimensionCombo, error) {
	allCoverage, err := s.diversityRepo.GetDimensionCoverage()
	if err != nil {
		return nil, err
	}

	// Shuffle and select random combinations
	rand.Shuffle(len(allCoverage), func(i, j int) {
		allCoverage[i], allCoverage[j] = allCoverage[j], allCoverage[i]
	})

	maxSelect := batchSize
	if len(allCoverage) < maxSelect {
		maxSelect = len(allCoverage)
	}

	combos := make([]models.DimensionCombo, 0, maxSelect)
	for i := 0; i < maxSelect; i++ {
		var combo models.DimensionCombo
		if err := combo.FromJSON(allCoverage[i].DimensionCombo); err != nil {
			continue
		}
		combos = append(combos, combo)
	}

	return combos, nil
}

// createPromptFromCombo creates a generation prompt based on dimension combination
func (s *DiversityService) createPromptFromCombo(combo models.DimensionCombo, config models.GenerationConfig) string {
	var promptParts []string

	// Add dimension constraints to prompt
	if combo.MealType != "" && combo.MealType != "なし" {
		promptParts = append(promptParts, fmt.Sprintf("食事タイプ: %s", combo.MealType))
	}
	if combo.Staple != "" && combo.Staple != "なし" {
		promptParts = append(promptParts, fmt.Sprintf("主食: %s", combo.Staple))
	}
	if combo.Protein != "" && combo.Protein != "なし" {
		promptParts = append(promptParts, fmt.Sprintf("メインの食材: %s", combo.Protein))
	}
	if combo.CookingMethod != "" {
		promptParts = append(promptParts, fmt.Sprintf("調理法: %s", combo.CookingMethod))
	}
	if combo.Seasoning != "" {
		promptParts = append(promptParts, fmt.Sprintf("味付け: %s", combo.Seasoning))
	}
	if combo.LazynessLevel != "" {
		lazynessDesc := map[string]string{
			"1_超簡単":   "超簡単で怠け者でも作れる",
			"2_簡単":    "簡単で手軽に作れる",
			"3_ちょい手間": "少し手間をかけても美味しい",
		}
		if desc, ok := lazynessDesc[combo.LazynessLevel]; ok {
			promptParts = append(promptParts, fmt.Sprintf("難易度: %s", desc))
		}
	}

	prompt := fmt.Sprintf(`以下の条件でレシピを作成してください：

%s

- 調理時間は15分以内
- 初心者でも作れる簡単なレシピ
- 日本の家庭でよく使われる食材
- 栄養バランスを考慮
- 手順は3ステップ以内`, strings.Join(promptParts, "\n"))

	return prompt
}

// generateSingleRecipe generates a single recipe (simplified implementation)
func (s *DiversityService) generateSingleRecipe(prompt string, config models.GenerationConfig) (*models.RecipeData, error) {
	// In a full implementation, this would use the actual generator service
	// For now, return a placeholder recipe

	// This would normally call s.generatorService.GenerateRecipe() with the prompt
	// but we'll return a simplified version for Phase 1

	placeholderRecipe := &models.RecipeData{
		Title:       fmt.Sprintf("多様化レシピ %d", time.Now().Unix()%1000),
		CookingTime: 10,
		Ingredients: []models.Ingredient{
			{Name: "基本食材", Amount: "適量"},
		},
		Steps:         []string{"調理する", "味付けする", "完成"},
		Tags:          []string{"簡単", "多様化"},
		Season:        "all",
		LazinessScore: 8.0,
		ServingSize:   1,
		Difficulty:    "easy",
	}

	return placeholderRecipe, nil
}

// InitializeSystem initializes the diversity system
func (s *DiversityService) InitializeSystem() error {
	log.Println("Initializing diversity system...")

	// Initialize dimension coverage combinations
	if err := s.diversityRepo.InitializeDimensionCoverage(); err != nil {
		return fmt.Errorf("failed to initialize dimension coverage: %w", err)
	}

	log.Println("Diversity system initialized successfully")
	return nil
}
