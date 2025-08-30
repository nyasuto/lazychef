package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"github.com/sashabaranov/go-openai"
	"lazychef/internal/models"
)

// EmbeddingDeduplicator handles recipe similarity detection using OpenAI embeddings
type EmbeddingDeduplicator struct {
	client              *openai.Client
	db                  *sql.DB
	embeddingVersion    string  // "v3"
	similarityThreshold float64 // cosine similarity threshold for duplicates
	jaccardThreshold    float64 // jaccard coefficient threshold for ingredients
}

// RecipeEmbedding represents a stored recipe embedding
type RecipeEmbedding struct {
	RecipeID         int       `json:"recipe_id"`
	EmbeddingVersion string    `json:"embedding_version"`
	ContentHash      string    `json:"content_hash"`
	Embedding        []float32 `json:"embedding"`
	Dimensions       int       `json:"dimensions"`
}

// DuplicateResult represents a detected duplicate
type DuplicateResult struct {
	RecipeID        int     `json:"recipe_id"`
	SimilarRecipeID int     `json:"similar_recipe_id"`
	SimilarityScore float64 `json:"similarity_score"`
	JaccardScore    float64 `json:"jaccard_score,omitempty"`
	DetectionMethod string  `json:"detection_method"`
}

// SimilarityReport contains results of a duplicate detection scan
type SimilarityReport struct {
	TotalRecipes    int               `json:"total_recipes"`
	ScannedRecipes  int               `json:"scanned_recipes"`
	DuplicatesFound int               `json:"duplicates_found"`
	Results         []DuplicateResult `json:"results"`
	ProcessingTime  string            `json:"processing_time"`
}

// NewEmbeddingDeduplicator creates a new embedding deduplicator
func NewEmbeddingDeduplicator(client *openai.Client, db *sql.DB) *EmbeddingDeduplicator {
	return &EmbeddingDeduplicator{
		client:              client,
		db:                  db,
		embeddingVersion:    "v3",
		similarityThreshold: 0.85, // 85% similarity threshold
		jaccardThreshold:    0.7,  // 70% ingredient overlap
	}
}

// ScanForDuplicates performs a full duplicate detection scan
func (d *EmbeddingDeduplicator) ScanForDuplicates(ctx context.Context, forceRefresh bool) (*SimilarityReport, error) {
	// Get all recipes
	recipes, err := d.getAllRecipes()
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}

	report := &SimilarityReport{
		TotalRecipes: len(recipes),
		Results:      []DuplicateResult{},
	}

	log.Printf("Starting duplicate scan for %d recipes", len(recipes))

	for i, recipe := range recipes {
		// Generate or update embedding if needed
		embedding, err := d.getOrCreateEmbedding(ctx, recipe, forceRefresh)
		if err != nil {
			log.Printf("Warning: failed to get embedding for recipe %d: %v", recipe.ID, err)
			continue
		}

		// Find similar recipes
		similarities, err := d.findSimilarRecipes(ctx, recipe, embedding)
		if err != nil {
			log.Printf("Warning: failed to find similarities for recipe %d: %v", recipe.ID, err)
			continue
		}

		// Add results
		for _, sim := range similarities {
			report.Results = append(report.Results, sim)

			// Save to database
			if err := d.saveDuplicateResult(sim); err != nil {
				log.Printf("Warning: failed to save duplicate result: %v", err)
			}
		}

		report.ScannedRecipes = i + 1

		// Log progress
		if (i+1)%10 == 0 {
			log.Printf("Processed %d/%d recipes", i+1, len(recipes))
		}
	}

	report.DuplicatesFound = len(report.Results)
	log.Printf("Duplicate scan completed: found %d potential duplicates out of %d recipes",
		report.DuplicatesFound, report.TotalRecipes)

	return report, nil
}

// CheckRecipeDuplicates checks if a specific recipe has duplicates
func (d *EmbeddingDeduplicator) CheckRecipeDuplicates(ctx context.Context, recipe *models.RecipeData) ([]DuplicateResult, error) {
	// Create temporary recipe object with ID 0 for checking
	tempRecipe := &RecipeWithID{
		ID:   0,
		Data: *recipe,
	}

	// Generate embedding
	embedding, err := d.generateEmbedding(ctx, recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Find similar recipes
	similarities, err := d.findSimilarRecipes(ctx, tempRecipe, &RecipeEmbedding{
		RecipeID:         0,
		EmbeddingVersion: d.embeddingVersion,
		Embedding:        embedding,
		Dimensions:       len(embedding),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find similar recipes: %w", err)
	}

	return similarities, nil
}

// RefreshEmbedding updates the embedding for a specific recipe
func (d *EmbeddingDeduplicator) RefreshEmbedding(ctx context.Context, recipeID int) error {
	recipe, err := d.getRecipeByID(recipeID)
	if err != nil {
		return fmt.Errorf("failed to get recipe: %w", err)
	}

	_, err = d.getOrCreateEmbedding(ctx, recipe, true)
	if err != nil {
		return fmt.Errorf("failed to refresh embedding: %w", err)
	}

	log.Printf("Refreshed embedding for recipe %d", recipeID)
	return nil
}

// GetDuplicateResults retrieves stored duplicate detection results
func (d *EmbeddingDeduplicator) GetDuplicateResults(limit int, method string) ([]DuplicateResult, error) {
	query := `
		SELECT recipe_id, similar_recipe_id, similarity_score, jaccard_score, detection_method
		FROM duplicate_detection_results 
		WHERE ($1 = '' OR detection_method = $1)
		ORDER BY similarity_score DESC, detected_at DESC
		LIMIT $2
	`

	rows, err := d.db.Query(query, method, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query duplicate results: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var results []DuplicateResult
	for rows.Next() {
		var result DuplicateResult
		var jaccardScore sql.NullFloat64

		err := rows.Scan(
			&result.RecipeID, &result.SimilarRecipeID, &result.SimilarityScore,
			&jaccardScore, &result.DetectionMethod,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan duplicate result: %w", err)
		}

		if jaccardScore.Valid {
			result.JaccardScore = jaccardScore.Float64
		}

		results = append(results, result)
	}

	return results, nil
}

// Helper methods

type RecipeWithID struct {
	ID   int               `json:"id"`
	Data models.RecipeData `json:"data"`
}

func (d *EmbeddingDeduplicator) getAllRecipes() ([]*RecipeWithID, error) {
	query := `SELECT id, data FROM recipes ORDER BY id`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query recipes: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var recipes []*RecipeWithID
	for rows.Next() {
		var recipe RecipeWithID
		var dataJSON string

		if err := rows.Scan(&recipe.ID, &dataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %w", err)
		}

		if err := json.Unmarshal([]byte(dataJSON), &recipe.Data); err != nil {
			log.Printf("Warning: failed to parse recipe %d: %v", recipe.ID, err)
			continue
		}

		recipes = append(recipes, &recipe)
	}

	return recipes, nil
}

func (d *EmbeddingDeduplicator) getRecipeByID(recipeID int) (*RecipeWithID, error) {
	query := `SELECT id, data FROM recipes WHERE id = ?`
	row := d.db.QueryRow(query, recipeID)

	var recipe RecipeWithID
	var dataJSON string

	if err := row.Scan(&recipe.ID, &dataJSON); err != nil {
		return nil, fmt.Errorf("failed to scan recipe: %w", err)
	}

	if err := json.Unmarshal([]byte(dataJSON), &recipe.Data); err != nil {
		return nil, fmt.Errorf("failed to parse recipe data: %w", err)
	}

	return &recipe, nil
}

func (d *EmbeddingDeduplicator) getOrCreateEmbedding(ctx context.Context, recipe *RecipeWithID, forceRefresh bool) (*RecipeEmbedding, error) {
	contentHash := d.generateContentHash(&recipe.Data)

	// Check if embedding exists and is current
	if !forceRefresh {
		if existing, err := d.getStoredEmbedding(recipe.ID); err == nil {
			if existing.ContentHash == contentHash && existing.EmbeddingVersion == d.embeddingVersion {
				return existing, nil
			}
		}
	}

	// Generate new embedding
	embeddingVector, err := d.generateEmbedding(ctx, &recipe.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	embedding := &RecipeEmbedding{
		RecipeID:         recipe.ID,
		EmbeddingVersion: d.embeddingVersion,
		ContentHash:      contentHash,
		Embedding:        embeddingVector,
		Dimensions:       len(embeddingVector),
	}

	// Store embedding
	if err := d.storeEmbedding(embedding); err != nil {
		return nil, fmt.Errorf("failed to store embedding: %w", err)
	}

	return embedding, nil
}

func (d *EmbeddingDeduplicator) generateEmbedding(ctx context.Context, recipe *models.RecipeData) ([]float32, error) {
	// Create text representation for embedding
	text := d.recipeToText(recipe)

	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2, // text-embedding-3-small or text-embedding-3-large
	}

	resp, err := d.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned from OpenAI")
	}

	return resp.Data[0].Embedding, nil
}

func (d *EmbeddingDeduplicator) recipeToText(recipe *models.RecipeData) string {
	var parts []string

	// Title (weighted heavily)
	parts = append(parts, fmt.Sprintf("Title: %s", recipe.Title))

	// Ingredients
	ingredientNames := make([]string, len(recipe.Ingredients))
	for i, ing := range recipe.Ingredients {
		ingredientNames[i] = ing.Name
	}
	parts = append(parts, fmt.Sprintf("Ingredients: %s", strings.Join(ingredientNames, ", ")))

	// Steps
	parts = append(parts, fmt.Sprintf("Steps: %s", strings.Join([]string(recipe.Steps), " ")))

	// Tags and season
	if len(recipe.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("Tags: %s", strings.Join(recipe.Tags, ", ")))
	}
	parts = append(parts, fmt.Sprintf("Season: %s", recipe.Season))

	return strings.Join(parts, "\n")
}

func (d *EmbeddingDeduplicator) generateContentHash(recipe *models.RecipeData) string {
	// Create deterministic hash of recipe content
	content := fmt.Sprintf("%s|%v|%v|%v|%s",
		recipe.Title, recipe.Ingredients, []string(recipe.Steps), recipe.Tags, recipe.Season)

	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func (d *EmbeddingDeduplicator) getStoredEmbedding(recipeID int) (*RecipeEmbedding, error) {
	query := `
		SELECT recipe_id, embedding_version, content_hash, embedding, dimensions
		FROM recipe_embeddings 
		WHERE recipe_id = ?
	`

	row := d.db.QueryRow(query, recipeID)

	var embedding RecipeEmbedding
	var embeddingBlob []byte

	err := row.Scan(
		&embedding.RecipeID, &embedding.EmbeddingVersion, &embedding.ContentHash,
		&embeddingBlob, &embedding.Dimensions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan embedding: %w", err)
	}

	// Deserialize embedding vector
	if err := json.Unmarshal(embeddingBlob, &embedding.Embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}

	return &embedding, nil
}

func (d *EmbeddingDeduplicator) storeEmbedding(embedding *RecipeEmbedding) error {
	// Serialize embedding vector
	embeddingBlob, err := json.Marshal(embedding.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO recipe_embeddings
		(recipe_id, embedding_version, content_hash, embedding, dimensions)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = d.db.Exec(query,
		embedding.RecipeID, embedding.EmbeddingVersion, embedding.ContentHash,
		embeddingBlob, embedding.Dimensions,
	)

	return err
}

func (d *EmbeddingDeduplicator) findSimilarRecipes(ctx context.Context, targetRecipe *RecipeWithID, targetEmbedding *RecipeEmbedding) ([]DuplicateResult, error) {
	// Get all other embeddings
	query := `
		SELECT recipe_id, embedding_version, content_hash, embedding, dimensions
		FROM recipe_embeddings 
		WHERE recipe_id != ? AND embedding_version = ?
	`

	rows, err := d.db.Query(query, targetRecipe.ID, d.embeddingVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var results []DuplicateResult

	for rows.Next() {
		var otherEmbedding RecipeEmbedding
		var embeddingBlob []byte

		err := rows.Scan(
			&otherEmbedding.RecipeID, &otherEmbedding.EmbeddingVersion,
			&otherEmbedding.ContentHash, &embeddingBlob, &otherEmbedding.Dimensions,
		)
		if err != nil {
			log.Printf("Warning: failed to scan embedding: %v", err)
			continue
		}

		// Deserialize embedding
		if err := json.Unmarshal(embeddingBlob, &otherEmbedding.Embedding); err != nil {
			log.Printf("Warning: failed to unmarshal embedding: %v", err)
			continue
		}

		// Calculate cosine similarity
		cosineSim := d.cosineSimilarity(targetEmbedding.Embedding, otherEmbedding.Embedding)

		if cosineSim >= d.similarityThreshold {
			// Get the other recipe for Jaccard calculation
			otherRecipe, err := d.getRecipeByID(otherEmbedding.RecipeID)
			if err != nil {
				log.Printf("Warning: failed to get recipe %d: %v", otherEmbedding.RecipeID, err)
				continue
			}

			// Calculate Jaccard coefficient for ingredients
			jaccardSim := d.jaccardSimilarity(targetRecipe.Data.Ingredients, otherRecipe.Data.Ingredients)

			// Determine detection method
			method := "embedding"
			if jaccardSim >= d.jaccardThreshold {
				method = "combined"
			}

			result := DuplicateResult{
				RecipeID:        targetRecipe.ID,
				SimilarRecipeID: otherEmbedding.RecipeID,
				SimilarityScore: cosineSim,
				JaccardScore:    jaccardSim,
				DetectionMethod: method,
			}

			results = append(results, result)
		}
	}

	// Sort by similarity score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	return results, nil
}

func (d *EmbeddingDeduplicator) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64

	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (d *EmbeddingDeduplicator) jaccardSimilarity(ingredientsA, ingredientsB []models.Ingredient) float64 {
	// Extract ingredient names and normalize
	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, ing := range ingredientsA {
		normalized := strings.ToLower(strings.TrimSpace(ing.Name))
		setA[normalized] = true
	}

	for _, ing := range ingredientsB {
		normalized := strings.ToLower(strings.TrimSpace(ing.Name))
		setB[normalized] = true
	}

	// Calculate intersection
	intersection := 0
	for ingredient := range setA {
		if setB[ingredient] {
			intersection++
		}
	}

	// Calculate union
	union := len(setA) + len(setB) - intersection

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func (d *EmbeddingDeduplicator) saveDuplicateResult(result DuplicateResult) error {
	query := `
		INSERT OR REPLACE INTO duplicate_detection_results
		(recipe_id, similar_recipe_id, similarity_score, jaccard_score, detection_method)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query,
		result.RecipeID, result.SimilarRecipeID, result.SimilarityScore,
		result.JaccardScore, result.DetectionMethod,
	)

	return err
}

// SetThresholds allows updating similarity thresholds
func (d *EmbeddingDeduplicator) SetThresholds(cosineThreshold, jaccardThreshold float64) {
	d.similarityThreshold = cosineThreshold
	d.jaccardThreshold = jaccardThreshold
}

// GetThresholds returns current similarity thresholds
func (d *EmbeddingDeduplicator) GetThresholds() (cosine, jaccard float64) {
	return d.similarityThreshold, d.jaccardThreshold
}
