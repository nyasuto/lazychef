package services

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"lazychef/internal/config"
	"lazychef/internal/models"
)

// BatchGenerationService handles OpenAI Batch API for cost-efficient mass generation
type BatchGenerationService struct {
	client           *openai.Client
	config           *config.OpenAIConfig
	db               *sql.DB
	batchStoragePath string
}

// GenerationJob represents a batch generation job
type GenerationJob struct {
	ID          string                `json:"id"`
	BatchType   string                `json:"batch_type"`
	Config      BatchGenerationConfig `json:"config"`
	ModelInfo   *ModelInfo            `json:"model_info,omitempty"`
	CostData    *CostData             `json:"cost_data,omitempty"`
	Status      string                `json:"status"`
	BatchID     string                `json:"batch_id,omitempty"`
	SubmittedAt *time.Time            `json:"submitted_at,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
}

// BatchGenerationConfig contains parameters for batch generation
type BatchGenerationConfig struct {
	Requests             []RecipeGenerationRequest `json:"requests"`
	Model                string                    `json:"model"`
	MaxTokens            int                       `json:"max_tokens"`
	Temperature          float32                   `json:"temperature,omitempty"`
	UseStructuredOutputs bool                      `json:"use_structured_outputs"`
	CompletionWindow     string                    `json:"completion_window"` // "24h" only for now
}

// ModelInfo stores model execution details
type ModelInfo struct {
	Model             string `json:"model"`
	Seed              *int   `json:"seed,omitempty"`
	SystemFingerprint string `json:"system_fingerprint,omitempty"`
}

// CostData tracks token usage and costs
type CostData struct {
	TotalTokens      int     `json:"total_tokens"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
	ActualCostUSD    float64 `json:"actual_cost_usd,omitempty"`
}

// BatchRequest represents a single request in the batch JSONL file
type BatchRequest struct {
	CustomID string                       `json:"custom_id"`
	Method   string                       `json:"method"`
	URL      string                       `json:"url"`
	Body     openai.ChatCompletionRequest `json:"body"`
}

// BatchResponse represents a single response from the batch output
type BatchResponse struct {
	ID       string                         `json:"id"`
	CustomID string                         `json:"custom_id"`
	Response *openai.ChatCompletionResponse `json:"response"`
	Error    *BatchError                    `json:"error"`
}

// BatchError represents an error in batch processing
type BatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewBatchGenerationService creates a new batch generation service
func NewBatchGenerationService(client *openai.Client, config *config.OpenAIConfig, db *sql.DB, storagePath string) *BatchGenerationService {
	return &BatchGenerationService{
		client:           client,
		config:           config,
		db:               db,
		batchStoragePath: storagePath,
	}
}

// SubmitBatchJob submits a new batch generation job
func (s *BatchGenerationService) SubmitBatchJob(ctx context.Context, config BatchGenerationConfig) (*GenerationJob, error) {
	jobID := uuid.New().String()

	job := &GenerationJob{
		ID:        jobID,
		BatchType: "batch_api",
		Config:    config,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Validate configuration
	if err := s.validateBatchConfig(config); err != nil {
		return nil, fmt.Errorf("invalid batch config: %w", err)
	}

	// Generate JSONL file
	inputFilePath, err := s.generateJSONLFile(job)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JSONL file: %w", err)
	}

	// Upload to OpenAI
	inputFile, err := s.uploadBatchFile(ctx, inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload batch file: %w", err)
	}

	// Create batch job
	batchReq := openai.CreateBatchRequest{
		InputFileID:      inputFile.ID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: config.CompletionWindow,
		Metadata: map[string]interface{}{
			"job_id":     jobID,
			"created_by": "lazychef",
		},
	}

	batch, err := s.client.CreateBatch(ctx, batchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	// Update job with batch ID
	job.BatchID = batch.ID
	job.Status = "submitted"
	now := time.Now()
	job.SubmittedAt = &now

	// Save to database
	if err := s.saveJob(job); err != nil {
		return nil, fmt.Errorf("failed to save job: %w", err)
	}

	log.Printf("Submitted batch job %s with OpenAI batch ID %s (%d requests)",
		jobID, batch.ID, len(config.Requests))

	return job, nil
}

// GetJobStatus retrieves the current status of a batch job
func (s *BatchGenerationService) GetJobStatus(ctx context.Context, jobID string) (*GenerationJob, error) {
	job, err := s.loadJob(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to load job: %w", err)
	}

	if job.BatchID == "" || job.Status == "completed" || job.Status == "failed" || job.Status == "cancelled" {
		return job, nil // No need to check OpenAI if no batch ID or already terminal
	}

	// Check status with OpenAI
	batch, err := s.client.RetrieveBatch(ctx, job.BatchID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch from OpenAI: %w", err)
	}

	// Update job status
	job.Status = batch.Status
	if batch.Status == "completed" {
		now := time.Now()
		job.CompletedAt = &now

		// Update cost data if available
		if batch.RequestCounts.Total > 0 {
			job.CostData = &CostData{
				TotalTokens: batch.RequestCounts.Total,
				// Note: Batch API doesn't provide detailed token breakdown
				EstimatedCostUSD: s.estimateBatchCost(batch.RequestCounts.Total, job.Config.Model),
			}
		}
	}

	// Save updated job
	if err := s.saveJob(job); err != nil {
		log.Printf("Warning: failed to save job update: %v", err)
	}

	return job, nil
}

// RetrieveBatchResults downloads and processes the results of a completed batch job
func (s *BatchGenerationService) RetrieveBatchResults(ctx context.Context, jobID string) ([]*models.RecipeData, error) {
	job, err := s.GetJobStatus(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	if job.Status != "completed" {
		return nil, fmt.Errorf("job not completed, current status: %s", job.Status)
	}

	// Get batch details from OpenAI
	batch, err := s.client.RetrieveBatch(ctx, job.BatchID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch: %w", err)
	}

	if batch.OutputFileID == nil || *batch.OutputFileID == "" {
		return nil, fmt.Errorf("no output file available")
	}

	// Download output file
	outputFilePath, err := s.downloadBatchOutput(ctx, *batch.OutputFileID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to download output: %w", err)
	}

	// Process results
	recipes, err := s.processBatchOutput(outputFilePath, job)
	if err != nil {
		return nil, fmt.Errorf("failed to process batch output: %w", err)
	}

	log.Printf("Retrieved %d recipes from batch job %s", len(recipes), jobID)
	return recipes, nil
}

// CancelBatchJob cancels a running batch job
func (s *BatchGenerationService) CancelBatchJob(ctx context.Context, jobID string) error {
	job, err := s.loadJob(jobID)
	if err != nil {
		return fmt.Errorf("failed to load job: %w", err)
	}

	if job.BatchID == "" {
		return fmt.Errorf("job has no batch ID")
	}

	// Cancel with OpenAI
	_, err = s.client.CancelBatch(ctx, job.BatchID)
	if err != nil {
		return fmt.Errorf("failed to cancel batch: %w", err)
	}

	// Update job status
	job.Status = "cancelled"
	if err := s.saveJob(job); err != nil {
		log.Printf("Warning: failed to save cancelled job: %v", err)
	}

	return nil
}

// ListJobs returns a list of batch jobs with optional filtering
func (s *BatchGenerationService) ListJobs(limit int, status string) ([]*GenerationJob, error) {
	query := `
		SELECT id, batch_type, config, model_info, cost_data, status, batch_id, 
		       submitted_at, completed_at, created_at 
		FROM recipe_generation_jobs 
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC 
		LIMIT $2
	`

	rows, err := s.db.Query(query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var jobs []*GenerationJob
	for rows.Next() {
		job := &GenerationJob{}
		var configJSON, modelInfoJSON, costDataJSON sql.NullString
		var submittedAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID, &job.BatchType, &configJSON, &modelInfoJSON, &costDataJSON,
			&job.Status, &job.BatchID, &submittedAt, &completedAt, &job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job row: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(configJSON.String), &job.Config); err != nil {
			log.Printf("Warning: failed to parse config for job %s: %v", job.ID, err)
		}

		if modelInfoJSON.Valid {
			if err := json.Unmarshal([]byte(modelInfoJSON.String), &job.ModelInfo); err != nil {
				log.Printf("Warning: failed to parse model info for job %s: %v", job.ID, err)
			}
		}

		if costDataJSON.Valid {
			if err := json.Unmarshal([]byte(costDataJSON.String), &job.CostData); err != nil {
				log.Printf("Warning: failed to parse cost data for job %s: %v", job.ID, err)
			}
		}

		if submittedAt.Valid {
			job.SubmittedAt = &submittedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Helper methods

func (s *BatchGenerationService) validateBatchConfig(config BatchGenerationConfig) error {
	if len(config.Requests) == 0 {
		return fmt.Errorf("no requests provided")
	}
	if len(config.Requests) > 50000 {
		return fmt.Errorf("too many requests, maximum 50000")
	}
	if config.Model == "" {
		return fmt.Errorf("model not specified")
	}
	if config.CompletionWindow != "24h" {
		return fmt.Errorf("completion_window must be '24h'")
	}
	return nil
}

func (s *BatchGenerationService) generateJSONLFile(job *GenerationJob) (string, error) {
	// Ensure storage directory exists
	if err := os.MkdirAll(s.batchStoragePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	filePath := filepath.Join(s.batchStoragePath, fmt.Sprintf("batch_%s_input.jsonl", job.ID))
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create JSONL file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close file: %v", err)
		}
	}()

	for i, req := range job.Config.Requests {
		customID := fmt.Sprintf("req_%s_%d", job.ID, i)

		// Generate prompt for this request
		promptTemplate := GetRecipeGenerationPrompt(req)

		// Build chat completion request
		chatReq := openai.ChatCompletionRequest{
			Model:     job.Config.Model,
			MaxTokens: job.Config.MaxTokens,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: promptTemplate.SystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: promptTemplate.UserPrompt,
				},
			},
		}

		// Add temperature only for non-GPT-5 models
		if !strings.HasPrefix(job.Config.Model, "gpt-5") {
			chatReq.Temperature = job.Config.Temperature
		}

		// Add Structured Outputs if enabled
		if job.Config.UseStructuredOutputs {
			schema := models.GetRecipeJSONSchema()
			schemaJSON, err := json.Marshal(schema)
			if err == nil {
				chatReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
					Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
					JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
						Name:   "recipe_with_safety",
						Schema: json.RawMessage(schemaJSON),
						Strict: true,
					},
				}
			}
		}

		batchReq := BatchRequest{
			CustomID: customID,
			Method:   "POST",
			URL:      "/v1/chat/completions",
			Body:     chatReq,
		}

		// Write as JSONL (one JSON object per line)
		if err := json.NewEncoder(file).Encode(batchReq); err != nil {
			return "", fmt.Errorf("failed to write batch request: %w", err)
		}
	}

	return filePath, nil
}

func (s *BatchGenerationService) uploadBatchFile(ctx context.Context, filePath string) (*openai.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close file: %v", err)
		}
	}()

	req := openai.FileRequest{
		FileName: filepath.Base(filePath),
		FilePath: filePath,
		Purpose:  "batch",
	}

	uploadedFile, err := s.client.CreateFile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &uploadedFile, nil
}

func (s *BatchGenerationService) downloadBatchOutput(ctx context.Context, fileID, jobID string) (string, error) {
	// Get file content from OpenAI
	content, err := s.client.GetFileContent(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		if err := content.Close(); err != nil {
			log.Printf("Warning: failed to close content: %v", err)
		}
	}()

	outputPath := filepath.Join(s.batchStoragePath, fmt.Sprintf("batch_%s_output.jsonl", jobID))
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Printf("Warning: failed to close output file: %v", err)
		}
	}()

	if _, err := io.Copy(outputFile, content); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return outputPath, nil
}

func (s *BatchGenerationService) processBatchOutput(filePath string, job *GenerationJob) ([]*models.RecipeData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close file: %v", err)
		}
	}()

	var recipes []*models.RecipeData
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var response BatchResponse
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			log.Printf("Warning: failed to parse batch response line: %v", err)
			continue
		}

		if response.Error != nil {
			log.Printf("Warning: batch request %s failed: %s", response.CustomID, response.Error.Message)
			continue
		}

		if response.Response == nil || len(response.Response.Choices) == 0 {
			log.Printf("Warning: no response for batch request %s", response.CustomID)
			continue
		}

		content := response.Response.Choices[0].Message.Content
		content = strings.TrimSpace(content)

		var recipe models.RecipeData
		if err := json.Unmarshal([]byte(content), &recipe); err != nil {
			log.Printf("Warning: failed to parse recipe JSON for %s: %v", response.CustomID, err)
			continue
		}

		recipes = append(recipes, &recipe)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan output file: %w", err)
	}

	return recipes, nil
}

func (s *BatchGenerationService) estimateBatchCost(totalRequests int, model string) float64 {
	// Rough cost estimation - this would need to be updated with actual pricing
	costPerRequest := 0.001 // $0.001 per request (example)
	if strings.Contains(model, "gpt-4") {
		costPerRequest = 0.005
	}

	return float64(totalRequests) * costPerRequest * 0.5 // 50% discount for batch API
}

func (s *BatchGenerationService) saveJob(job *GenerationJob) error {
	configJSON, err := json.Marshal(job.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var modelInfoJSON, costDataJSON sql.NullString

	if job.ModelInfo != nil {
		data, err := json.Marshal(job.ModelInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal model info: %w", err)
		}
		modelInfoJSON = sql.NullString{String: string(data), Valid: true}
	}

	if job.CostData != nil {
		data, err := json.Marshal(job.CostData)
		if err != nil {
			return fmt.Errorf("failed to marshal cost data: %w", err)
		}
		costDataJSON = sql.NullString{String: string(data), Valid: true}
	}

	query := `
		INSERT OR REPLACE INTO recipe_generation_jobs 
		(id, batch_type, config, model_info, cost_data, status, batch_id, submitted_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		job.ID, job.BatchType, string(configJSON), modelInfoJSON, costDataJSON,
		job.Status, job.BatchID, job.SubmittedAt, job.CompletedAt, job.CreatedAt,
	)

	return err
}

func (s *BatchGenerationService) loadJob(jobID string) (*GenerationJob, error) {
	query := `
		SELECT id, batch_type, config, model_info, cost_data, status, batch_id,
		       submitted_at, completed_at, created_at
		FROM recipe_generation_jobs 
		WHERE id = ?
	`

	row := s.db.QueryRow(query, jobID)

	job := &GenerationJob{}
	var configJSON, modelInfoJSON, costDataJSON sql.NullString
	var submittedAt, completedAt sql.NullTime

	err := row.Scan(
		&job.ID, &job.BatchType, &configJSON, &modelInfoJSON, &costDataJSON,
		&job.Status, &job.BatchID, &submittedAt, &completedAt, &job.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan job: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(configJSON.String), &job.Config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if modelInfoJSON.Valid {
		if err := json.Unmarshal([]byte(modelInfoJSON.String), &job.ModelInfo); err != nil {
			return nil, fmt.Errorf("failed to parse model info: %w", err)
		}
	}

	if costDataJSON.Valid {
		if err := json.Unmarshal([]byte(costDataJSON.String), &job.CostData); err != nil {
			return nil, fmt.Errorf("failed to parse cost data: %w", err)
		}
	}

	if submittedAt.Valid {
		job.SubmittedAt = &submittedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return job, nil
}
