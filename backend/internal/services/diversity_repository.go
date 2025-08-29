package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"lazychef/internal/database"
	"lazychef/internal/models"
)

// DiversityRepository handles diversity system database operations
type DiversityRepository struct {
	db *database.Database
}

// NewDiversityRepository creates a new diversity repository
func NewDiversityRepository(db *database.Database) *DiversityRepository {
	return &DiversityRepository{
		db: db,
	}
}

// GetRecipeDimensions retrieves all recipe dimensions
func (r *DiversityRepository) GetRecipeDimensions() ([]*models.RecipeDimension, error) {
	query := `
		SELECT id, dimension_type, dimension_value, weight, is_active, created_at
		FROM recipe_dimensions
		WHERE is_active = 1
		ORDER BY dimension_type, weight DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query recipe dimensions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	dimensions := make([]*models.RecipeDimension, 0)
	for rows.Next() {
		var dimension models.RecipeDimension
		if err := rows.Scan(
			&dimension.ID,
			&dimension.DimensionType,
			&dimension.DimensionValue,
			&dimension.Weight,
			&dimension.IsActive,
			&dimension.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dimension: %w", err)
		}
		dimensions = append(dimensions, &dimension)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dimension rows: %w", err)
	}

	return dimensions, nil
}

// GetDimensionsByType retrieves dimensions by type
func (r *DiversityRepository) GetDimensionsByType(dimensionType string) ([]*models.RecipeDimension, error) {
	query := `
		SELECT id, dimension_type, dimension_value, weight, is_active, created_at
		FROM recipe_dimensions
		WHERE dimension_type = ? AND is_active = 1
		ORDER BY weight DESC
	`

	rows, err := r.db.Query(query, dimensionType)
	if err != nil {
		return nil, fmt.Errorf("failed to query dimensions by type: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	dimensions := make([]*models.RecipeDimension, 0)
	for rows.Next() {
		var dimension models.RecipeDimension
		if err := rows.Scan(
			&dimension.ID,
			&dimension.DimensionType,
			&dimension.DimensionValue,
			&dimension.Weight,
			&dimension.IsActive,
			&dimension.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dimension: %w", err)
		}
		dimensions = append(dimensions, &dimension)
	}

	return dimensions, nil
}

// GetDimensionCoverage retrieves coverage data
func (r *DiversityRepository) GetDimensionCoverage() ([]*models.DimensionCoverage, error) {
	query := `
		SELECT id, dimension_combo, current_count, target_count, priority_score,
		       last_generated_at, created_at, updated_at
		FROM dimension_coverage
		ORDER BY priority_score ASC, current_count ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query dimension coverage: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	coverages := make([]*models.DimensionCoverage, 0)
	for rows.Next() {
		var coverage models.DimensionCoverage
		var lastGenerated sql.NullTime

		if err := rows.Scan(
			&coverage.ID,
			&coverage.DimensionCombo,
			&coverage.CurrentCount,
			&coverage.TargetCount,
			&coverage.PriorityScore,
			&lastGenerated,
			&coverage.CreatedAt,
			&coverage.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan coverage: %w", err)
		}

		if lastGenerated.Valid {
			coverage.LastGeneratedAt = &lastGenerated.Time
		}

		coverages = append(coverages, &coverage)
	}

	return coverages, nil
}

// GetLowCoverageCombinations retrieves combinations with low coverage
func (r *DiversityRepository) GetLowCoverageCombinations(limit int) ([]*models.DimensionCoverage, error) {
	query := `
		SELECT id, dimension_combo, current_count, target_count, priority_score,
		       last_generated_at, created_at, updated_at
		FROM dimension_coverage
		WHERE current_count < target_count
		ORDER BY priority_score ASC, current_count ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query low coverage combinations: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	coverages := make([]*models.DimensionCoverage, 0)
	for rows.Next() {
		var coverage models.DimensionCoverage
		var lastGenerated sql.NullTime

		if err := rows.Scan(
			&coverage.ID,
			&coverage.DimensionCombo,
			&coverage.CurrentCount,
			&coverage.TargetCount,
			&coverage.PriorityScore,
			&lastGenerated,
			&coverage.CreatedAt,
			&coverage.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan low coverage: %w", err)
		}

		if lastGenerated.Valid {
			coverage.LastGeneratedAt = &lastGenerated.Time
		}

		coverages = append(coverages, &coverage)
	}

	return coverages, nil
}

// UpsertDimensionCoverage creates or updates coverage data
func (r *DiversityRepository) UpsertDimensionCoverage(combo string, increment int) error {
	query := `
		INSERT INTO dimension_coverage (dimension_combo, current_count, last_generated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(dimension_combo) DO UPDATE SET
			current_count = current_count + ?,
			last_generated_at = CURRENT_TIMESTAMP,
			priority_score = CASE 
				WHEN (current_count + ?) < target_count 
				THEN (target_count - (current_count + ?)) / CAST(target_count AS REAL)
				ELSE 0.1
			END
	`

	err := r.db.Execute(query, combo, increment, increment, increment, increment)
	if err != nil {
		return fmt.Errorf("failed to upsert dimension coverage: %w", err)
	}

	return nil
}

// GetGenerationProfile retrieves a generation profile by name
func (r *DiversityRepository) GetGenerationProfile(name string) (*models.GenerationProfile, error) {
	query := `
		SELECT id, profile_name, config, performance_data, is_active, created_at, updated_at
		FROM generation_profiles
		WHERE profile_name = ? AND is_active = 1
	`

	var profile models.GenerationProfile
	var configStr, perfDataStr string
	row := r.db.QueryRow(query, name)

	if err := row.Scan(
		&profile.ID,
		&profile.ProfileName,
		&configStr,
		&perfDataStr,
		&profile.IsActive,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to query generation profile: %w", err)
	}

	// Convert strings to json.RawMessage
	profile.Config = json.RawMessage(configStr)
	profile.PerformanceData = json.RawMessage(perfDataStr)

	return &profile, nil
}

// UpdateProfilePerformance updates performance data for a profile
func (r *DiversityRepository) UpdateProfilePerformance(profileName string, perfData *models.PerformanceData) error {
	perfDataJSON, err := json.Marshal(perfData)
	if err != nil {
		return fmt.Errorf("failed to marshal performance data: %w", err)
	}

	query := `
		UPDATE generation_profiles
		SET performance_data = ?, updated_at = CURRENT_TIMESTAMP
		WHERE profile_name = ?
	`

	err = r.db.Execute(query, string(perfDataJSON), profileName)
	if err != nil {
		return fmt.Errorf("failed to update profile performance: %w", err)
	}

	return nil
}

// InitializeDimensionCoverage creates initial coverage entries for all dimension combinations
func (r *DiversityRepository) InitializeDimensionCoverage() error {
	// Get all dimension types and values
	dimensions, err := r.GetRecipeDimensions()
	if err != nil {
		return fmt.Errorf("failed to get dimensions: %w", err)
	}

	// Group by dimension type
	dimensionMap := make(map[string][]string)
	for _, dim := range dimensions {
		dimensionMap[dim.DimensionType] = append(dimensionMap[dim.DimensionType], dim.DimensionValue)
	}

	// Required dimension types
	requiredTypes := []string{"meal_type", "staple", "protein", "cooking_method", "seasoning", "laziness_level"}

	// Verify all required types exist
	for _, reqType := range requiredTypes {
		if _, exists := dimensionMap[reqType]; !exists {
			return fmt.Errorf("missing required dimension type: %s", reqType)
		}
	}

	// Generate all possible combinations
	count := 0
	for _, mealType := range dimensionMap["meal_type"] {
		for _, staple := range dimensionMap["staple"] {
			for _, protein := range dimensionMap["protein"] {
				for _, cookingMethod := range dimensionMap["cooking_method"] {
					for _, seasoning := range dimensionMap["seasoning"] {
						for _, lazynessLevel := range dimensionMap["laziness_level"] {
							combo := models.DimensionCombo{
								MealType:      mealType,
								Staple:        staple,
								Protein:       protein,
								CookingMethod: cookingMethod,
								Seasoning:     seasoning,
								LazynessLevel: lazynessLevel,
							}

							comboJSON, err := combo.ToJSON()
							if err != nil {
								continue // Skip invalid combinations
							}

							// Calculate priority based on dimension weights
							priority := r.calculateComboPriority(dimensions, combo)

							query := `
								INSERT OR IGNORE INTO dimension_coverage 
								(dimension_combo, current_count, target_count, priority_score)
								VALUES (?, 0, ?, ?)
							`

							err = r.db.Execute(query, comboJSON, 3, priority) // Default target: 3 recipes per combo
							if err != nil {
								log.Printf("Warning: failed to insert coverage for combo %s: %v", comboJSON, err)
								continue
							}
							count++
						}
					}
				}
			}
		}
	}

	log.Printf("Initialized %d dimension coverage combinations", count)
	return nil
}

// calculateComboPriority calculates priority score based on dimension weights
func (r *DiversityRepository) calculateComboPriority(dimensions []*models.RecipeDimension, combo models.DimensionCombo) float64 {
	weightMap := make(map[string]map[string]float64)

	for _, dim := range dimensions {
		if weightMap[dim.DimensionType] == nil {
			weightMap[dim.DimensionType] = make(map[string]float64)
		}
		weightMap[dim.DimensionType][dim.DimensionValue] = dim.Weight
	}

	// Calculate weighted priority (higher weights = lower priority score)
	totalWeight := 0.0
	totalWeight += weightMap["meal_type"][combo.MealType]
	totalWeight += weightMap["staple"][combo.Staple]
	totalWeight += weightMap["protein"][combo.Protein]
	totalWeight += weightMap["cooking_method"][combo.CookingMethod]
	totalWeight += weightMap["seasoning"][combo.Seasoning]
	totalWeight += weightMap["laziness_level"][combo.LazynessLevel]

	// Invert weight so higher weight = lower priority score (higher priority)
	maxPossibleWeight := 2.0 * 6 // Assuming max weight is 2.0 per dimension
	priority := (maxPossibleWeight - totalWeight) / maxPossibleWeight

	// Ensure priority is between 0.1 and 1.0
	if priority < 0.1 {
		priority = 0.1
	}
	if priority > 1.0 {
		priority = 1.0
	}

	return priority
}
