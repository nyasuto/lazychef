package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"lazychef/internal/database"
	"lazychef/internal/models"
)

// RecipeQualityService handles quality assessment for generated recipes
type RecipeQualityService struct {
	db                    *database.Database
	diversityService      *DiversityService
	embeddingDeduplicator *EmbeddingDeduplicator
}

// QualityScore represents the quality assessment of a recipe
type QualityScore struct {
	OverallScore       float64                   `json:"overall_score"` // 0-100
	DetailedScores     map[string]float64        `json:"detailed_scores"`
	DimensionAlignment map[string]AlignmentScore `json:"dimension_alignment"`
	QualityIssues      []QualityIssue            `json:"quality_issues"`
	Recommendations    []string                  `json:"recommendations"`
	PassesThreshold    bool                      `json:"passes_threshold"`
}

// AlignmentScore represents how well a recipe aligns with its assigned dimensions
type AlignmentScore struct {
	DimensionType  string   `json:"dimension_type"`
	DimensionValue string   `json:"dimension_value"`
	AlignmentScore float64  `json:"alignment_score"` // 0-100
	Confidence     float64  `json:"confidence"`      // 0-1
	Issues         []string `json:"issues,omitempty"`
}

// QualityIssue represents a specific quality problem
type QualityIssue struct {
	Severity    string  `json:"severity"` // critical, major, minor
	Category    string  `json:"category"` // completeness, consistency, feasibility, etc.
	Description string  `json:"description"`
	Impact      float64 `json:"impact"` // Score reduction
}

// QualityReport represents a comprehensive quality analysis
type QualityReport struct {
	TotalRecipes        int                `json:"total_recipes"`
	AverageQuality      float64            `json:"average_quality"`
	PassingRecipes      int                `json:"passing_recipes"`
	FailingRecipes      int                `json:"failing_recipes"`
	QualityDistribution map[string]int     `json:"quality_distribution"`
	CommonIssues        map[string]int     `json:"common_issues"`
	DimensionAccuracy   map[string]float64 `json:"dimension_accuracy"`
	GeneratedAt         time.Time          `json:"generated_at"`
}

// NewRecipeQualityService creates a new quality service
func NewRecipeQualityService(db *database.Database, diversityService *DiversityService, embeddingDeduplicator *EmbeddingDeduplicator) *RecipeQualityService {
	return &RecipeQualityService{
		db:                    db,
		diversityService:      diversityService,
		embeddingDeduplicator: embeddingDeduplicator,
	}
}

// AssessRecipeQuality performs comprehensive quality assessment on a recipe
func (s *RecipeQualityService) AssessRecipeQuality(recipe *models.RecipeData, dimensions []DimensionCombination) (*QualityScore, error) {
	score := &QualityScore{
		DetailedScores:     make(map[string]float64),
		DimensionAlignment: make(map[string]AlignmentScore),
		QualityIssues:      []QualityIssue{},
		Recommendations:    []string{},
	}

	// 1. Completeness Check (25% weight)
	completenessScore := s.assessCompleteness(recipe, score)
	score.DetailedScores["completeness"] = completenessScore

	// 2. Consistency Check (20% weight)
	consistencyScore := s.assessConsistency(recipe, score)
	score.DetailedScores["consistency"] = consistencyScore

	// 3. Feasibility Check (20% weight)
	feasibilityScore := s.assessFeasibility(recipe, score)
	score.DetailedScores["feasibility"] = feasibilityScore

	// 4. Dimension Alignment (20% weight)
	alignmentScore := s.assessDimensionAlignment(recipe, dimensions, score)
	score.DetailedScores["dimension_alignment"] = alignmentScore

	// 5. Laziness Score Accuracy (15% weight)
	lazinessAccuracy := s.assessLazinessAccuracy(recipe, score)
	score.DetailedScores["laziness_accuracy"] = lazinessAccuracy

	// Calculate overall score with weights
	score.OverallScore = completenessScore*0.25 +
		consistencyScore*0.20 +
		feasibilityScore*0.20 +
		alignmentScore*0.20 +
		lazinessAccuracy*0.15

	// Determine if passes quality threshold (70%)
	score.PassesThreshold = score.OverallScore >= 70.0

	// Generate recommendations
	s.generateRecommendations(score)

	return score, nil
}

// assessCompleteness checks if recipe has all required components
func (s *RecipeQualityService) assessCompleteness(recipe *models.RecipeData, score *QualityScore) float64 {
	completenessScore := 100.0

	// Check title
	if recipe.Title == "" {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "critical",
			Category:    "completeness",
			Description: "Recipe title is missing",
			Impact:      30.0,
		})
		completenessScore -= 30.0
	} else if len(recipe.Title) < 5 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "major",
			Category:    "completeness",
			Description: "Recipe title is too short",
			Impact:      10.0,
		})
		completenessScore -= 10.0
	}

	// Check ingredients
	if len(recipe.Ingredients) == 0 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "critical",
			Category:    "completeness",
			Description: "No ingredients specified",
			Impact:      30.0,
		})
		completenessScore -= 30.0
	} else if len(recipe.Ingredients) < 3 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "major",
			Category:    "completeness",
			Description: fmt.Sprintf("Too few ingredients (%d)", len(recipe.Ingredients)),
			Impact:      15.0,
		})
		completenessScore -= 15.0
	}

	// Check steps
	if len(recipe.Steps) == 0 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "critical",
			Category:    "completeness",
			Description: "No cooking steps provided",
			Impact:      30.0,
		})
		completenessScore -= 30.0
	} else if len(recipe.Steps) > 3 {
		// LazyChef requirement: maximum 3 steps
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "major",
			Category:    "completeness",
			Description: fmt.Sprintf("Too many steps (%d) for lazy cooking", len(recipe.Steps)),
			Impact:      20.0,
		})
		completenessScore -= 20.0
	}

	// Check cooking time
	if recipe.CookingTime <= 0 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "major",
			Category:    "completeness",
			Description: "Invalid cooking time",
			Impact:      10.0,
		})
		completenessScore -= 10.0
	}

	return math.Max(0, completenessScore)
}

// assessConsistency checks internal consistency of recipe
func (s *RecipeQualityService) assessConsistency(recipe *models.RecipeData, score *QualityScore) float64 {
	consistencyScore := 100.0

	// Check if ingredients mentioned in steps
	for _, ingredient := range recipe.Ingredients {
		mentioned := false
		for _, step := range recipe.Steps {
			if strings.Contains(strings.ToLower(step), strings.ToLower(ingredient.Name)) {
				mentioned = true
				break
			}
		}
		if !mentioned {
			score.QualityIssues = append(score.QualityIssues, QualityIssue{
				Severity:    "minor",
				Category:    "consistency",
				Description: fmt.Sprintf("Ingredient '%s' not mentioned in steps", ingredient.Name),
				Impact:      5.0,
			})
			consistencyScore -= 5.0
		}
	}

	// Check cooking time vs steps complexity
	estimatedTime := len(recipe.Steps) * 5 // Rough estimate: 5 min per step
	if recipe.CookingTime > 0 {
		timeDifference := math.Abs(float64(recipe.CookingTime - estimatedTime))
		if timeDifference > 10 {
			score.QualityIssues = append(score.QualityIssues, QualityIssue{
				Severity:    "minor",
				Category:    "consistency",
				Description: "Cooking time doesn't match step complexity",
				Impact:      10.0,
			})
			consistencyScore -= 10.0
		}
	}

	// Check laziness score consistency
	if recipe.LazinessScore > 8 && len(recipe.Steps) > 2 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "minor",
			Category:    "consistency",
			Description: "High laziness score but multiple steps",
			Impact:      10.0,
		})
		consistencyScore -= 10.0
	}

	return math.Max(0, consistencyScore)
}

// assessFeasibility checks if recipe is practical to make
func (s *RecipeQualityService) assessFeasibility(recipe *models.RecipeData, score *QualityScore) float64 {
	feasibilityScore := 100.0

	// Check cooking time feasibility
	if recipe.CookingTime > 30 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "major",
			Category:    "feasibility",
			Description: fmt.Sprintf("Cooking time too long for lazy cooking (%d min)", recipe.CookingTime),
			Impact:      20.0,
		})
		feasibilityScore -= 20.0
	}

	// Check ingredient availability
	rareIngredients := 0
	for _, ingredient := range recipe.Ingredients {
		if s.isRareIngredient(ingredient.Name) {
			rareIngredients++
		}
	}
	if rareIngredients > 2 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "minor",
			Category:    "feasibility",
			Description: fmt.Sprintf("Contains %d rare/specialty ingredients", rareIngredients),
			Impact:      15.0,
		})
		feasibilityScore -= 15.0
	}

	// Check step complexity
	for _, step := range recipe.Steps {
		if len(step) > 200 {
			score.QualityIssues = append(score.QualityIssues, QualityIssue{
				Severity:    "minor",
				Category:    "feasibility",
				Description: "Step description too complex",
				Impact:      10.0,
			})
			feasibilityScore -= 10.0
			break
		}
	}

	return math.Max(0, feasibilityScore)
}

// assessDimensionAlignment checks if recipe matches assigned dimensions
func (s *RecipeQualityService) assessDimensionAlignment(recipe *models.RecipeData, dimensions []DimensionCombination, score *QualityScore) float64 {
	if len(dimensions) == 0 {
		return 100.0 // No dimensions to check
	}

	alignmentScore := 100.0
	combo := dimensions[0] // Use first dimension combination

	// Check meal type alignment
	if combo.MealType != nil {
		alignment := s.checkMealTypeAlignment(recipe, combo.MealType.DimensionValue)
		score.DimensionAlignment["meal_type"] = AlignmentScore{
			DimensionType:  "meal_type",
			DimensionValue: combo.MealType.DimensionValue,
			AlignmentScore: alignment,
			Confidence:     0.8,
		}
		if alignment < 70 {
			alignmentScore -= 20.0
		}
	}

	// Check protein alignment
	if combo.Protein != nil {
		alignment := s.checkProteinAlignment(recipe, combo.Protein.DimensionValue)
		score.DimensionAlignment["protein"] = AlignmentScore{
			DimensionType:  "protein",
			DimensionValue: combo.Protein.DimensionValue,
			AlignmentScore: alignment,
			Confidence:     0.9,
		}
		if alignment < 70 {
			alignmentScore -= 20.0
		}
	}

	// Check cooking method alignment
	if combo.CookingMethod != nil {
		alignment := s.checkCookingMethodAlignment(recipe, combo.CookingMethod.DimensionValue)
		score.DimensionAlignment["cooking_method"] = AlignmentScore{
			DimensionType:  "cooking_method",
			DimensionValue: combo.CookingMethod.DimensionValue,
			AlignmentScore: alignment,
			Confidence:     0.85,
		}
		if alignment < 70 {
			alignmentScore -= 20.0
		}
	}

	return math.Max(0, alignmentScore)
}

// assessLazinessAccuracy checks if laziness score is accurate
func (s *RecipeQualityService) assessLazinessAccuracy(recipe *models.RecipeData, score *QualityScore) float64 {
	// Calculate expected laziness score
	expectedScore := 10.0

	// Reduce based on cooking time
	if recipe.CookingTime > 15 {
		expectedScore -= 2.0
	}
	if recipe.CookingTime > 30 {
		expectedScore -= 3.0
	}

	// Reduce based on steps
	expectedScore -= float64(len(recipe.Steps)-1) * 1.5

	// Reduce based on ingredients
	if len(recipe.Ingredients) > 5 {
		expectedScore -= 1.0
	}

	expectedScore = math.Max(1.0, math.Min(10.0, expectedScore))

	// Calculate accuracy
	difference := math.Abs(recipe.LazinessScore - expectedScore)
	accuracy := 100.0 - (difference * 10.0)

	if difference > 2 {
		score.QualityIssues = append(score.QualityIssues, QualityIssue{
			Severity:    "minor",
			Category:    "accuracy",
			Description: fmt.Sprintf("Laziness score mismatch (expected: %.1f, got: %.1f)", expectedScore, recipe.LazinessScore),
			Impact:      15.0,
		})
	}

	return math.Max(0, accuracy)
}

// Helper methods
func (s *RecipeQualityService) isRareIngredient(name string) bool {
	rareIngredients := []string{
		"トリュフ", "キャビア", "フォアグラ", "サフラン",
		"松茸", "雲丹", "いくら", "あわび",
	}

	nameLower := strings.ToLower(name)
	for _, rare := range rareIngredients {
		if strings.Contains(nameLower, strings.ToLower(rare)) {
			return true
		}
	}
	return false
}

func (s *RecipeQualityService) checkMealTypeAlignment(recipe *models.RecipeData, mealType string) float64 {
	// Simple keyword-based alignment check
	keywords := map[string][]string{
		"主菜":     {"肉", "魚", "メイン", "焼", "煮", "炒"},
		"副菜":     {"サラダ", "和え", "おひたし", "煮物", "野菜"},
		"汁物":     {"スープ", "味噌汁", "汁", "だし"},
		"主食":     {"ご飯", "パン", "麺", "パスタ", "うどん"},
		"おやつ・甘味": {"ケーキ", "クッキー", "プリン", "甘", "デザート"},
	}

	alignmentScore := 0.0
	if targetKeywords, exists := keywords[mealType]; exists {
		for _, keyword := range targetKeywords {
			if strings.Contains(recipe.Title, keyword) {
				alignmentScore += 30.0
			}
			for _, step := range recipe.Steps {
				if strings.Contains(step, keyword) {
					alignmentScore += 10.0
					break
				}
			}
		}
	}

	return math.Min(100.0, alignmentScore)
}

func (s *RecipeQualityService) checkProteinAlignment(recipe *models.RecipeData, protein string) float64 {
	// Check if specified protein is in ingredients
	alignmentScore := 0.0
	proteinLower := strings.ToLower(protein)

	for _, ingredient := range recipe.Ingredients {
		if strings.Contains(strings.ToLower(ingredient.Name), proteinLower) {
			alignmentScore = 100.0
			break
		}
	}

	// Also check in title
	if strings.Contains(strings.ToLower(recipe.Title), proteinLower) {
		alignmentScore = math.Max(alignmentScore, 80.0)
	}

	return alignmentScore
}

func (s *RecipeQualityService) checkCookingMethodAlignment(recipe *models.RecipeData, method string) float64 {
	// Check if cooking method is mentioned in steps
	alignmentScore := 0.0
	methodLower := strings.ToLower(method)

	for _, step := range recipe.Steps {
		if strings.Contains(strings.ToLower(step), methodLower) {
			alignmentScore = 100.0
			break
		}
	}

	// Partial credit for related methods
	relatedMethods := map[string][]string{
		"電子レンジ": {"レンジ", "温め", "チン"},
		"炒める":   {"炒め", "フライパン", "油"},
		"煮る":    {"煮", "鍋", "だし", "スープ"},
		"焼く":    {"焼", "オーブン", "グリル"},
		"和えるだけ": {"和え", "混ぜ", "かける"},
	}

	if related, exists := relatedMethods[method]; exists {
		for _, keyword := range related {
			for _, step := range recipe.Steps {
				if strings.Contains(strings.ToLower(step), strings.ToLower(keyword)) {
					alignmentScore = math.Max(alignmentScore, 70.0)
				}
			}
		}
	}

	return alignmentScore
}

func (s *RecipeQualityService) generateRecommendations(score *QualityScore) {
	// Generate recommendations based on issues
	if score.OverallScore < 50 {
		score.Recommendations = append(score.Recommendations,
			"Consider regenerating this recipe with stricter parameters")
	}

	// Check for critical issues
	criticalCount := 0
	for _, issue := range score.QualityIssues {
		if issue.Severity == "critical" {
			criticalCount++
		}
	}
	if criticalCount > 0 {
		score.Recommendations = append(score.Recommendations,
			fmt.Sprintf("Address %d critical issues before using this recipe", criticalCount))
	}

	// Dimension alignment recommendations
	for _, alignment := range score.DimensionAlignment {
		if alignment.AlignmentScore < 50 {
			score.Recommendations = append(score.Recommendations,
				fmt.Sprintf("Improve %s alignment (current: %.0f%%)",
					alignment.DimensionType, alignment.AlignmentScore))
		}
	}

	// Completeness recommendations
	if compScore, exists := score.DetailedScores["completeness"]; exists && compScore < 70 {
		score.Recommendations = append(score.Recommendations,
			"Ensure all required recipe components are present")
	}
}

// GenerateQualityReport generates a comprehensive quality report for multiple recipes
func (s *RecipeQualityService) GenerateQualityReport(recipes []models.Recipe, dimensionMappings map[int][]DimensionCombination) (*QualityReport, error) {
	report := &QualityReport{
		TotalRecipes:        len(recipes),
		QualityDistribution: make(map[string]int),
		CommonIssues:        make(map[string]int),
		DimensionAccuracy:   make(map[string]float64),
		GeneratedAt:         time.Now(),
	}

	totalQuality := 0.0
	dimensionScores := make(map[string][]float64)

	for _, recipe := range recipes {
		dimensions := dimensionMappings[recipe.ID]
		qualityScore, err := s.AssessRecipeQuality(&recipe.Data, dimensions)
		if err != nil {
			continue
		}

		// Update totals
		totalQuality += qualityScore.OverallScore
		if qualityScore.PassesThreshold {
			report.PassingRecipes++
		} else {
			report.FailingRecipes++
		}

		// Categorize quality
		category := s.categorizeQuality(qualityScore.OverallScore)
		report.QualityDistribution[category]++

		// Track common issues
		for _, issue := range qualityScore.QualityIssues {
			report.CommonIssues[issue.Category]++
		}

		// Track dimension accuracy
		for dimType, alignment := range qualityScore.DimensionAlignment {
			dimensionScores[dimType] = append(dimensionScores[dimType], alignment.AlignmentScore)
		}
	}

	// Calculate averages
	if report.TotalRecipes > 0 {
		report.AverageQuality = totalQuality / float64(report.TotalRecipes)
	}

	// Calculate dimension accuracy averages
	for dimType, scores := range dimensionScores {
		if len(scores) > 0 {
			total := 0.0
			for _, score := range scores {
				total += score
			}
			report.DimensionAccuracy[dimType] = total / float64(len(scores))
		}
	}

	return report, nil
}

func (s *RecipeQualityService) categorizeQuality(score float64) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 75:
		return "good"
	case score >= 60:
		return "acceptable"
	case score >= 40:
		return "poor"
	default:
		return "unacceptable"
	}
}
