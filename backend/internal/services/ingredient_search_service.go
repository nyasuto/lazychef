package services

import (
	"database/sql"
	"fmt"
	"strings"
)

// IngredientSearchService handles hierarchical ingredient searching
type IngredientSearchService struct {
	db *sql.DB
}

// NewIngredientSearchService creates a new ingredient search service
func NewIngredientSearchService(db *sql.DB) *IngredientSearchService {
	return &IngredientSearchService{db: db}
}

// IngredientSearchResult represents the result of ingredient search
type IngredientSearchResult struct {
	SpecificIngredients []string `json:"specific_ingredients"`
	MatchedGroups       []string `json:"matched_groups"`
}

// SearchIngredientsByGroup searches for specific ingredients by group name or display name
// This function handles the hierarchical ingredient classification system
func (s *IngredientSearchService) SearchIngredientsByGroup(searchTerms []string) (*IngredientSearchResult, error) {
	if len(searchTerms) == 0 {
		return &IngredientSearchResult{
			SpecificIngredients: []string{},
			MatchedGroups:       []string{},
		}, nil
	}

	var allSpecificIngredients []string
	var matchedGroups []string

	for _, term := range searchTerms {
		// Clean the search term
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		// First, try exact match in specific ingredients
		specificIngredients, err := s.findSpecificIngredients(term)
		if err != nil {
			return nil, fmt.Errorf("specific ingredient search failed: %w", err)
		}

		// If found specific ingredients, add them directly
		if len(specificIngredients) > 0 {
			allSpecificIngredients = append(allSpecificIngredients, specificIngredients...)
			continue
		}

		// If not found in specific ingredients, search in groups
		groupIngredients, groups, err := s.findIngredientsByGroup(term)
		if err != nil {
			return nil, fmt.Errorf("group ingredient search failed: %w", err)
		}

		allSpecificIngredients = append(allSpecificIngredients, groupIngredients...)
		matchedGroups = append(matchedGroups, groups...)
	}

	// Remove duplicates
	allSpecificIngredients = removeDuplicates(allSpecificIngredients)
	matchedGroups = removeDuplicates(matchedGroups)

	return &IngredientSearchResult{
		SpecificIngredients: allSpecificIngredients,
		MatchedGroups:       matchedGroups,
	}, nil
}

// findSpecificIngredients searches for ingredients by exact name or aliases
func (s *IngredientSearchService) findSpecificIngredients(searchTerm string) ([]string, error) {
	query := `
		SELECT DISTINCT name 
		FROM specific_ingredients 
		WHERE name = ? 
		   OR display_name = ?
		   OR (aliases IS NOT NULL AND json_extract(aliases, '$') LIKE ?)
	`

	aliasPattern := fmt.Sprintf(`%%"%s"%%`, searchTerm)
	rows, err := s.db.Query(query, searchTerm, searchTerm, aliasPattern)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't fail the function
			_ = closeErr
		}
	}()

	var ingredients []string
	for rows.Next() {
		var ingredient string
		if err := rows.Scan(&ingredient); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

// findIngredientsByGroup searches for ingredients by group name (hierarchical)
func (s *IngredientSearchService) findIngredientsByGroup(searchTerm string) ([]string, []string, error) {
	// Search in ingredient groups (both name and display_name)
	query := `
		SELECT DISTINCT si.name, ig.display_name
		FROM specific_ingredients si
		JOIN ingredient_group_mappings igm ON si.id = igm.ingredient_id
		JOIN ingredient_groups ig ON igm.group_id = ig.id
		WHERE ig.name = ? 
		   OR ig.display_name = ?
		   OR ig.name LIKE ?
		   OR ig.display_name LIKE ?
		ORDER BY si.name
	`

	likePattern := "%" + searchTerm + "%"
	rows, err := s.db.Query(query, searchTerm, searchTerm, likePattern, likePattern)
	if err != nil {
		return nil, nil, fmt.Errorf("group query execution failed: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't fail the function
			_ = closeErr
		}
	}()

	var ingredients []string
	var groups []string
	groupSet := make(map[string]bool) // To avoid duplicate groups

	for rows.Next() {
		var ingredient, groupName string
		if err := rows.Scan(&ingredient, &groupName); err != nil {
			return nil, nil, fmt.Errorf("group row scan failed: %w", err)
		}
		ingredients = append(ingredients, ingredient)
		if !groupSet[groupName] {
			groups = append(groups, groupName)
			groupSet[groupName] = true
		}
	}

	return ingredients, groups, nil
}

// GetAvailableGroups returns all ingredient groups for UI dropdown
func (s *IngredientSearchService) GetAvailableGroups() ([]IngredientGroup, error) {
	query := `
		SELECT id, name, display_name, parent_id, level, sort_order
		FROM ingredient_groups
		ORDER BY level, sort_order, display_name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("group query failed: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't fail the function
			_ = closeErr
		}
	}()

	var groups []IngredientGroup
	for rows.Next() {
		var group IngredientGroup
		var parentID sql.NullInt64

		if err := rows.Scan(&group.ID, &group.Name, &group.DisplayName,
			&parentID, &group.Level, &group.SortOrder); err != nil {
			return nil, fmt.Errorf("group scan failed: %w", err)
		}

		if parentID.Valid {
			group.ParentID = &parentID.Int64
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// IngredientGroup represents an ingredient group for API response
type IngredientGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	ParentID    *int64 `json:"parent_id,omitempty"`
	Level       int    `json:"level"`
	SortOrder   int    `json:"sort_order"`
}

// GetHierarchicalGroups returns groups organized by hierarchy
func (s *IngredientSearchService) GetHierarchicalGroups() (map[string]interface{}, error) {
	groups, err := s.GetAvailableGroups()
	if err != nil {
		return nil, err
	}

	// Organize into hierarchy
	hierarchy := make(map[string]interface{})
	level1Groups := make(map[int64]IngredientGroup)
	level2Groups := make(map[int64][]IngredientGroup)

	// First pass: collect level 1 groups
	for _, group := range groups {
		if group.Level == 1 {
			level1Groups[group.ID] = group
		}
	}

	// Second pass: collect level 2 groups under their parents
	for _, group := range groups {
		if group.Level == 2 && group.ParentID != nil {
			parentID := *group.ParentID
			level2Groups[parentID] = append(level2Groups[parentID], group)
		}
	}

	// Build final hierarchy
	for _, parentGroup := range level1Groups {
		children := level2Groups[parentGroup.ID]
		hierarchy[parentGroup.DisplayName] = map[string]interface{}{
			"id":           parentGroup.ID,
			"name":         parentGroup.Name,
			"display_name": parentGroup.DisplayName,
			"level":        parentGroup.Level,
			"children":     children,
		}
	}

	return hierarchy, nil
}

// BuildIngredientSearchCondition builds SQL condition for ingredient search
// This replaces the simple LIKE search with hierarchical search
func (s *IngredientSearchService) BuildIngredientSearchCondition(ingredients []string) (string, []interface{}, error) {
	if len(ingredients) == 0 {
		return "", []interface{}{}, nil
	}

	// Get all specific ingredients that match the search terms
	searchResult, err := s.SearchIngredientsByGroup(ingredients)
	if err != nil {
		return "", nil, fmt.Errorf("ingredient search failed: %w", err)
	}

	// Combine original search terms with found specific ingredients
	allIngredients := append(ingredients, searchResult.SpecificIngredients...)
	allIngredients = removeDuplicates(allIngredients)

	if len(allIngredients) == 0 {
		return "", []interface{}{}, nil
	}

	// Build SQL condition
	var conditions []string
	var args []interface{}

	for _, ingredient := range allIngredients {
		conditions = append(conditions, `EXISTS (
			SELECT 1 FROM json_each(json_extract(data, '$.ingredients'))
			WHERE json_extract(value, '$.name') = ?
		)`)
		args = append(args, ingredient)
	}

	sqlCondition := " AND (" + strings.Join(conditions, " OR ") + ")"
	return sqlCondition, args, nil
}

// Helper function to remove duplicates from string slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
