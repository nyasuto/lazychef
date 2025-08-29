package models

import (
	"encoding/json"
	"time"
)

// RecipeDimension represents a single dimension for recipe categorization
type RecipeDimension struct {
	ID             int       `json:"id" db:"id"`
	DimensionType  string    `json:"dimension_type" db:"dimension_type"`
	DimensionValue string    `json:"dimension_value" db:"dimension_value"`
	Weight         float64   `json:"weight" db:"weight"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// DimensionCoverage tracks recipe coverage for dimension combinations
type DimensionCoverage struct {
	ID              int        `json:"id" db:"id"`
	DimensionCombo  string     `json:"dimension_combo" db:"dimension_combo"`
	CurrentCount    int        `json:"current_count" db:"current_count"`
	TargetCount     int        `json:"target_count" db:"target_count"`
	PriorityScore   float64    `json:"priority_score" db:"priority_score"`
	LastGeneratedAt *time.Time `json:"last_generated_at" db:"last_generated_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// GenerationProfile defines parameters for diverse recipe generation
type GenerationProfile struct {
	ID              int             `json:"id" db:"id"`
	ProfileName     string          `json:"profile_name" db:"profile_name"`
	Config          json.RawMessage `json:"config" db:"config"`
	PerformanceData json.RawMessage `json:"performance_data" db:"performance_data"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// DimensionCombo represents a parsed dimension combination
type DimensionCombo struct {
	MealType      string `json:"meal_type"`
	Staple        string `json:"staple"`
	Protein       string `json:"protein"`
	CookingMethod string `json:"cooking_method"`
	Seasoning     string `json:"seasoning"`
	LazynessLevel string `json:"laziness_level"`
}

// ToJSON converts DimensionCombo to JSON string
func (dc *DimensionCombo) ToJSON() (string, error) {
	data, err := json.Marshal(dc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses JSON string into DimensionCombo
func (dc *DimensionCombo) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), dc)
}

// GenerationConfig defines parameters for diverse generation
type GenerationConfig struct {
	Strategy         string             `json:"strategy"` // "coverage_first", "priority_first", "random_sample"
	BatchSize        int                `json:"batch_size"`
	MaxSimilarity    float64            `json:"max_similarity"`
	QualityThreshold float64            `json:"quality_threshold"`
	DimensionWeights map[string]float64 `json:"dimension_weights,omitempty"`
	FocusDimensions  []string           `json:"focus_dimensions,omitempty"`
}

// PerformanceData tracks generation performance metrics
type PerformanceData struct {
	SuccessRate      float64    `json:"success_rate"`
	AvgCostPerRecipe float64    `json:"avg_cost_per_recipe"`
	TotalGenerated   int        `json:"total_generated"`
	LastUpdated      *time.Time `json:"last_updated,omitempty"`
}

// CoverageAnalysis provides coverage metrics
type CoverageAnalysis struct {
	TotalCombinations   int                      `json:"total_combinations"`
	CoveredCombinations int                      `json:"covered_combinations"`
	CoverageRate        float64                  `json:"coverage_rate"`
	LowCoverageCombos   []CoverageSummary        `json:"low_coverage_combos"`
	DimensionStats      map[string]DimensionStat `json:"dimension_stats"`
}

// CoverageSummary summarizes coverage for a specific combination
type CoverageSummary struct {
	Combo        DimensionCombo `json:"combo"`
	CurrentCount int            `json:"current_count"`
	TargetCount  int            `json:"target_count"`
	Priority     float64        `json:"priority"`
	Gap          int            `json:"gap"`
}

// DimensionStat provides statistics for a dimension type
type DimensionStat struct {
	DimensionType string  `json:"dimension_type"`
	TotalValues   int     `json:"total_values"`
	CoveredValues int     `json:"covered_values"`
	AvgCoverage   float64 `json:"avg_coverage"`
}

// DiverseGenerationRequest defines parameters for diverse generation
type DiverseGenerationRequest struct {
	ProfileName      string             `json:"profile_name" binding:"required"`
	BatchSize        int                `json:"batch_size" binding:"min=1,max=50"`
	Strategy         string             `json:"strategy,omitempty"`
	MaxSimilarity    *float64           `json:"max_similarity,omitempty"`
	QualityThreshold *float64           `json:"quality_threshold,omitempty"`
	ForceDimensions  map[string]string  `json:"force_dimensions,omitempty"`
	CustomWeights    map[string]float64 `json:"custom_weights,omitempty"`
}

// DiverseGenerationResponse returns generation results with diversity metrics
type DiverseGenerationResponse struct {
	JobID          string         `json:"job_id"`
	ProfileUsed    string         `json:"profile_used"`
	Strategy       string         `json:"strategy"`
	RequestedCount int            `json:"requested_count"`
	GeneratedCount int            `json:"generated_count"`
	DiversityScore float64        `json:"diversity_score"`
	CoverageImpact CoverageImpact `json:"coverage_impact"`
	EstimatedCost  float64        `json:"estimated_cost"`
	Recipes        []RecipeData   `json:"recipes,omitempty"`
}

// CoverageImpact shows how generation affected coverage
type CoverageImpact struct {
	NewCombinations      int `json:"new_combinations"`
	ImprovedCombinations int `json:"improved_combinations"`
	TotalCombinations    int `json:"total_combinations"`
}

// Validate validates the diverse generation request
func (r *DiverseGenerationRequest) Validate() error {
	if r.ProfileName == "" {
		return ErrInvalidProfileName
	}
	if r.BatchSize < 1 || r.BatchSize > 50 {
		return ErrInvalidBatchSize
	}
	if r.Strategy != "" && r.Strategy != "coverage_first" && r.Strategy != "priority_first" && r.Strategy != "random_sample" {
		return ErrInvalidStrategy
	}
	if r.MaxSimilarity != nil && (*r.MaxSimilarity < 0.0 || *r.MaxSimilarity > 1.0) {
		return ErrInvalidSimilarity
	}
	if r.QualityThreshold != nil && (*r.QualityThreshold < 1.0 || *r.QualityThreshold > 10.0) {
		return ErrInvalidQualityThreshold
	}
	return nil
}
