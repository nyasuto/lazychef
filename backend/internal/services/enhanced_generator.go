package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"lazychef/internal/config"
	"lazychef/internal/models"
)

// EnhancedRecipeGeneratorService provides GPT-5 compatible recipe generation with Structured Outputs
type EnhancedRecipeGeneratorService struct {
	client              *openai.Client
	config              *config.OpenAIConfig
	rateLimiter         *RateLimiter
	cache               *RecipeCache
	foodSafetyValidator *FoodSafetyValidator
	qualityValidator    *QualityCheckService
}

// GenerationStage represents the stage of recipe generation
type GenerationStage string

const (
	StageIdeation  GenerationStage = "ideation"
	StageAuthoring GenerationStage = "authoring"
	StageCritique  GenerationStage = "critique"
)

// EnhancedGenerationRequest includes stage-specific parameters
type EnhancedGenerationRequest struct {
	RecipeGenerationRequest
	Stage           GenerationStage `json:"stage"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	Verbosity       string          `json:"verbosity,omitempty"`
	Seed            *int            `json:"seed,omitempty"`
}

// EnhancedGenerationResult includes detailed generation metadata
type EnhancedGenerationResult struct {
	*GenerationResult
	Stage              GenerationStage     `json:"stage"`
	ModelUsed          string              `json:"model_used"`
	ReasoningEffort    string              `json:"reasoning_effort"`
	Verbosity          string              `json:"verbosity"`
	Seed               *int                `json:"seed,omitempty"`
	SystemFingerprint  string              `json:"system_fingerprint,omitempty"`
	SafetyCheckResult  *SafetyCheckResult  `json:"safety_check_result,omitempty"`
	QualityCheckResult *QualityCheckResult `json:"quality_check_result,omitempty"`
	StructuredOutputs  bool                `json:"structured_outputs"`
}

// NewEnhancedRecipeGeneratorService creates a new enhanced generator service
func NewEnhancedRecipeGeneratorService(client *openai.Client, config *config.OpenAIConfig, rateLimiter *RateLimiter, cache *RecipeCache) *EnhancedRecipeGeneratorService {
	return &EnhancedRecipeGeneratorService{
		client:              client,
		config:              config,
		rateLimiter:         rateLimiter,
		cache:               cache,
		foodSafetyValidator: NewFoodSafetyValidator(config.FoodSafetyStrictMode),
		qualityValidator:    NewQualityCheckService(config),
	}
}

// GetConfig returns the OpenAI config
func (s *EnhancedRecipeGeneratorService) GetConfig() *config.OpenAIConfig {
	return s.config
}

// GetFoodSafetyValidator returns the food safety validator
func (s *EnhancedRecipeGeneratorService) GetFoodSafetyValidator() *FoodSafetyValidator {
	return s.foodSafetyValidator
}

// GetQualityValidator returns the quality validator
func (s *EnhancedRecipeGeneratorService) GetQualityValidator() *QualityCheckService {
	return s.qualityValidator
}

// GenerateRecipeEnhanced generates a recipe using GPT-5 with comprehensive validation
func (s *EnhancedRecipeGeneratorService) GenerateRecipeEnhanced(ctx context.Context, req EnhancedGenerationRequest) (*EnhancedGenerationResult, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	// Check cache first
	cacheKey := s.generateEnhancedCacheKey(req)
	if cachedResult := s.cache.Get(cacheKey); cachedResult != nil {
		enhancedResult := &EnhancedGenerationResult{
			GenerationResult:  cachedResult,
			Stage:             req.Stage,
			StructuredOutputs: s.config.UseStructuredOutputs,
		}
		enhancedResult.Metadata.CacheHit = true
		enhancedResult.Metadata.RequestID = requestID
		enhancedResult.Metadata.ProcessingTime = time.Since(startTime)
		return enhancedResult, nil
	}

	// Rate limiting
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting error: %w", err)
	}

	// Select appropriate model based on stage
	model := s.selectModelForStage(req.Stage)

	// Generate recipe using selected model
	recipe, tokensUsed, systemFingerprint, err := s.generateWithStructuredOutputs(ctx, req, model)
	if err != nil {
		return &EnhancedGenerationResult{
			GenerationResult: &GenerationResult{
				Metadata: GenerationMetadata{
					RequestID:      requestID,
					Model:          model,
					GeneratedAt:    time.Now(),
					ProcessingTime: time.Since(startTime),
					CacheHit:       false,
				},
				Error: err.Error(),
			},
			Stage:             req.Stage,
			ModelUsed:         model,
			StructuredOutputs: s.config.UseStructuredOutputs,
		}, err
	}

	// Perform food safety validation
	safetyResult, err := s.foodSafetyValidator.ValidateRecipe(recipe)
	if err != nil {
		return nil, fmt.Errorf("safety validation failed: %w", err)
	}

	// Perform quality validation
	qualityResult, err := s.qualityValidator.ValidateRecipe(recipe)
	if err != nil {
		return nil, fmt.Errorf("quality validation failed: %w", err)
	}

	// Create enhanced result
	result := &EnhancedGenerationResult{
		GenerationResult: &GenerationResult{
			Recipe: recipe,
			Metadata: GenerationMetadata{
				RequestID:      requestID,
				Model:          model,
				GeneratedAt:    time.Now(),
				TokensUsed:     tokensUsed,
				ProcessingTime: time.Since(startTime),
				CacheHit:       false,
			},
		},
		Stage:              req.Stage,
		ModelUsed:          model,
		ReasoningEffort:    req.ReasoningEffort,
		Verbosity:          req.Verbosity,
		Seed:               req.Seed,
		SystemFingerprint:  systemFingerprint,
		SafetyCheckResult:  safetyResult,
		QualityCheckResult: qualityResult,
		StructuredOutputs:  s.config.UseStructuredOutputs,
	}

	// Check if generation should be rejected based on safety/quality
	if s.config.FoodSafetyStrictMode && !safetyResult.Passed {
		result.Error = fmt.Sprintf("Recipe failed safety validation: %v", safetyResult.Violations)
		return result, fmt.Errorf("recipe failed safety validation")
	}

	// Cache successful result
	s.cache.Set(cacheKey, result.GenerationResult)

	log.Printf("Enhanced recipe generation completed: stage=%s, model=%s, safety_passed=%t, quality_score=%.2f",
		req.Stage, model, safetyResult.Passed, qualityResult.OverallScore)

	return result, nil
}

// generateWithStructuredOutputs calls OpenAI API with Structured Outputs if enabled
func (s *EnhancedRecipeGeneratorService) generateWithStructuredOutputs(ctx context.Context, req EnhancedGenerationRequest, model string) (*models.RecipeData, int, string, error) {
	// Generate enhanced prompt with food safety instructions
	prompt := s.generateEnhancedPrompt(req)

	// Build request
	chatReq := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt.SystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt.UserPrompt,
			},
		},
	}

	// Set temperature only for non-GPT-5 models (GPT-5 has beta limitations)
	if !s.isGPT5Model(model) {
		chatReq.Temperature = s.config.Temperature
	}

	// Add GPT-5 specific parameters
	if s.isGPT5Model(model) {
		// Set reasoning effort and verbosity for GPT-5
		// Note: ReasoningEffort and Verbosity parameters not yet supported in go-openai library
		// Will be enabled when library support is available:
		// if req.ReasoningEffort != "" {
		//     chatReq.ReasoningEffort = req.ReasoningEffort
		// }
		// if req.Verbosity != "" {
		//     chatReq.Verbosity = req.Verbosity
		// }
		if req.Seed != nil {
			chatReq.Seed = req.Seed
		}
	}

	// Add Structured Outputs if enabled
	if s.config.UseStructuredOutputs {
		schema := models.GetRecipeJSONSchema()

		// Debug: print schema to see what's being sent
		schemaJSON, err := json.MarshalIndent(schema, "", "  ")
		if err == nil {
			log.Printf("Generated JSON Schema: %s", string(schemaJSON))
		}

		chatReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "recipe_with_safety",
				Schema: json.RawMessage(schemaJSON),
				Strict: true,
			},
		}
	}

	// Set completion token limit for GPT-5 models
	if s.isGPT5Model(model) && s.config.MaxCompletionTokens > 0 {
		// GPT-5 models use MaxCompletionTokens instead of MaxTokens
		// Note: This would be the correct parameter when supported
		// chatReq.MaxCompletionTokens = s.config.MaxCompletionTokens
	} else if s.config.MaxTokens > 0 {
		// Legacy models use MaxTokens
		chatReq.MaxTokens = s.config.MaxTokens
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	// Call OpenAI API
	resp, err := s.client.CreateChatCompletion(timeoutCtx, chatReq)
	if err != nil {
		return nil, 0, "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, resp.Usage.TotalTokens, resp.SystemFingerprint, errors.New("no choices returned from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Parse JSON response
	var recipe models.RecipeData
	if err := json.Unmarshal([]byte(content), &recipe); err != nil {
		return nil, resp.Usage.TotalTokens, resp.SystemFingerprint, fmt.Errorf("failed to parse recipe JSON: %w", err)
	}

	return &recipe, resp.Usage.TotalTokens, resp.SystemFingerprint, nil
}

// selectModelForStage returns the appropriate model for the generation stage
func (s *EnhancedRecipeGeneratorService) selectModelForStage(stage GenerationStage) string {
	switch stage {
	case StageIdeation:
		return s.config.IdeationModel
	case StageAuthoring:
		return s.config.AuthoringModel
	case StageCritique:
		return s.config.CritiqueModel
	default:
		return s.config.AuthoringModel // Default to authoring model
	}
}

// isGPT5Model checks if the model is a GPT-5 variant
func (s *EnhancedRecipeGeneratorService) isGPT5Model(model string) bool {
	return strings.HasPrefix(model, "gpt-5")
}

// generateEnhancedPrompt creates an enhanced prompt with food safety instructions
func (s *EnhancedRecipeGeneratorService) generateEnhancedPrompt(req EnhancedGenerationRequest) PromptTemplate {
	basePrompt := GetRecipeGenerationPrompt(req.RecipeGenerationRequest)

	// Add food safety instructions to system prompt
	safetyInstructions := `

CRITICAL FOOD SAFETY REQUIREMENTS:
- Always specify safe cooking temperatures for meat, poultry, and eggs
- Chicken/poultry: 165°F, Ground beef/pork: 160°F, Whole cuts beef/pork/lamb: 145°F
- Never suggest eating raw flour, raw eggs in no-bake recipes, or undercooked meat
- Include allergen warnings for common allergens (nuts, dairy, eggs, wheat, soy, fish, shellfish)
- Avoid dangerous practices: thawing on counter, reusing marinades, rinsing raw chicken

LAZYCHEF CONSTRAINTS:
- Maximum 3 cooking steps for ultimate laziness
- Total time ≤ 15 minutes preferred
- Focus on simple, accessible ingredients
- Minimize active cooking time
- Single pan/pot cooking when possible`

	enhancedSystemPrompt := basePrompt.SystemPrompt + safetyInstructions

	// Add structured output instruction if enabled
	if s.config.UseStructuredOutputs {
		enhancedSystemPrompt += "\n\nIMPORTANT: Respond with valid JSON matching the exact schema provided. Include the safety_compliance object with required food safety information."
	}

	return PromptTemplate{
		SystemPrompt: enhancedSystemPrompt,
		UserPrompt:   basePrompt.UserPrompt,
	}
}

// generateEnhancedCacheKey generates a cache key for enhanced requests
func (s *EnhancedRecipeGeneratorService) generateEnhancedCacheKey(req EnhancedGenerationRequest) string {
	baseKey := s.generateCacheKey(req.RecipeGenerationRequest)
	return fmt.Sprintf("%s:stage=%s:structured=%t", baseKey, req.Stage, s.config.UseStructuredOutputs)
}

// generateCacheKey generates a cache key from the base request
func (s *EnhancedRecipeGeneratorService) generateCacheKey(req RecipeGenerationRequest) string {
	return fmt.Sprintf("recipe:%s:%d:%s", strings.Join(req.Ingredients, ","), req.MaxCookingTime, strings.Join(req.Preferences, ","))
}
