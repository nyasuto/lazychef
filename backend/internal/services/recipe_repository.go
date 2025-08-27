package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"lazychef/internal/database"
	"lazychef/internal/models"
	"log"
)

// RecipeRepository handles recipe database operations
type RecipeRepository struct {
	db *database.Database
}

// NewRecipeRepository creates a new recipe repository
func NewRecipeRepository(db *database.Database) *RecipeRepository {
	return &RecipeRepository{
		db: db,
	}
}

// SaveRecipe saves a recipe to the database
func (r *RecipeRepository) SaveRecipe(recipe *models.Recipe) error {
	data, err := json.Marshal(recipe)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe: %w", err)
	}

	query := `
		INSERT INTO recipes (data)
		VALUES (?)
	`

	err = r.db.Execute(query, string(data))
	if err != nil {
		return fmt.Errorf("failed to insert recipe: %w", err)
	}

	// Get last insert ID using separate query
	id, err := r.db.GetLastInsertID()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	recipe.ID = int(id)
	return nil
}

// GetRecipe retrieves a recipe by ID
func (r *RecipeRepository) GetRecipe(id int) (*models.Recipe, error) {
	query := `
		SELECT data FROM recipes WHERE id = ?
	`

	var data string
	row := r.db.QueryRow(query, id)
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to query recipe: %w", err)
	}

	var recipe models.Recipe
	if err := json.Unmarshal([]byte(data), &recipe); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe: %w", err)
	}

	recipe.ID = id
	return &recipe, nil
}

// SearchRecipes searches for recipes based on criteria
func (r *RecipeRepository) SearchRecipes(criteria models.SearchCriteria) ([]*models.Recipe, error) {
	query := `
		SELECT id, data FROM recipes
		WHERE 1=1
	`

	args := []interface{}{}

	// Add search conditions
	if criteria.Tag != "" {
		query += ` AND json_extract(data, '$.tags') LIKE ?`
		args = append(args, "%"+criteria.Tag+"%")
	}

	if criteria.Ingredient != "" {
		query += ` AND json_extract(data, '$.ingredients') LIKE ?`
		args = append(args, "%"+criteria.Ingredient+"%")
	}

	if criteria.MaxCookingTime > 0 {
		query += ` AND cooking_time <= ?`
		args = append(args, criteria.MaxCookingTime)
	}

	if criteria.MinLazinessScore > 0 {
		query += ` AND laziness_score >= ?`
		args = append(args, criteria.MinLazinessScore)
	}

	if criteria.Season != "" && criteria.Season != "all" {
		query += ` AND (season = ? OR season = 'all')`
		args = append(args, criteria.Season)
	}

	// Add ordering and limit
	query += ` ORDER BY created_at DESC`

	if criteria.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, criteria.Limit)
	} else {
		query += ` LIMIT 20` // Default limit
	}

	if criteria.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, criteria.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recipes: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	recipes := make([]*models.Recipe, 0)
	for rows.Next() {
		var id int
		var data string

		if err := rows.Scan(&id, &data); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var recipe models.Recipe
		if err := json.Unmarshal([]byte(data), &recipe); err != nil {
			continue // Skip invalid recipes
		}

		recipe.ID = id
		recipes = append(recipes, &recipe)
	}

	return recipes, nil
}

// GetRandomRecipes gets random recipes
func (r *RecipeRepository) GetRandomRecipes(count int) ([]*models.Recipe, error) {
	query := `
		SELECT id, data FROM recipes
		ORDER BY RANDOM()
		LIMIT ?
	`

	rows, err := r.db.Query(query, count)
	if err != nil {
		return nil, fmt.Errorf("failed to query random recipes: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	recipes := make([]*models.Recipe, 0, count)
	for rows.Next() {
		var id int
		var data string

		if err := rows.Scan(&id, &data); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var recipe models.Recipe
		if err := json.Unmarshal([]byte(data), &recipe); err != nil {
			continue
		}

		recipe.ID = id
		recipes = append(recipes, &recipe)
	}

	return recipes, nil
}

// CountRecipes counts total recipes in database
func (r *RecipeRepository) CountRecipes() (int, error) {
	query := `SELECT COUNT(*) FROM recipes`

	var count int
	row := r.db.QueryRow(query)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count recipes: %w", err)
	}

	return count, nil
}
