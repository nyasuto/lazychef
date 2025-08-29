package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"lazychef/internal/models"
	"lazychef/internal/services"
)

// AdminHandler handles administrative endpoints for batch processing and analytics
type AdminHandler struct {
	batchService     *services.BatchGenerationService
	embeddingService *services.EmbeddingDeduplicator
	tokenRateLimiter *services.TokenRateLimiter
	diversityService *services.DiversityService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(batchService *services.BatchGenerationService, embeddingService *services.EmbeddingDeduplicator, tokenRateLimiter *services.TokenRateLimiter, diversityService *services.DiversityService) *AdminHandler {
	return &AdminHandler{
		batchService:     batchService,
		embeddingService: embeddingService,
		tokenRateLimiter: tokenRateLimiter,
		diversityService: diversityService,
	}
}

// Batch Generation Endpoints

// SubmitBatchGeneration submits a new batch generation job
// POST /api/admin/batch-generation/submit
func (h *AdminHandler) SubmitBatchGeneration(c *gin.Context) {
	var config services.BatchGenerationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Default values
	if config.CompletionWindow == "" {
		config.CompletionWindow = "24h"
	}
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}

	job, err := h.batchService.SubmitBatchJob(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to submit batch job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"job_id":         job.ID,
			"batch_id":       job.BatchID,
			"status":         job.Status,
			"requests_count": len(config.Requests),
			"submitted_at":   job.SubmittedAt,
		},
	})
}

// GetBatchStatus gets the status of a batch job
// GET /api/admin/batch-generation/status/:job_id
func (h *AdminHandler) GetBatchStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Job ID is required",
		})
		return
	}

	job, err := h.batchService.GetJobStatus(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Job not found",
			"details": err.Error(),
		})
		return
	}

	response := gin.H{
		"success": true,
		"data": gin.H{
			"job_id":       job.ID,
			"batch_id":     job.BatchID,
			"status":       job.Status,
			"batch_type":   job.BatchType,
			"created_at":   job.CreatedAt,
			"submitted_at": job.SubmittedAt,
			"completed_at": job.CompletedAt,
		},
	}

	if job.CostData != nil {
		response["data"].(gin.H)["cost_data"] = job.CostData
	}

	c.JSON(http.StatusOK, response)
}

// CancelBatchJob cancels a running batch job
// POST /api/admin/batch-generation/cancel/:job_id
func (h *AdminHandler) CancelBatchJob(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Job ID is required",
		})
		return
	}

	err := h.batchService.CancelBatchJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cancel job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Batch job cancelled successfully",
	})
}

// GetBatchResults retrieves results from a completed batch job
// GET /api/admin/batch-generation/results/:job_id
func (h *AdminHandler) GetBatchResults(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Job ID is required",
		})
		return
	}

	recipes, err := h.batchService.RetrieveBatchResults(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve results",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"job_id":       jobID,
			"recipe_count": len(recipes),
			"recipes":      recipes,
		},
	})
}

// ListBatchJobs lists batch jobs with optional filtering
// GET /api/admin/batch-generation/jobs
func (h *AdminHandler) ListBatchJobs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	jobs, err := h.batchService.ListJobs(limit, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list jobs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"jobs":  jobs,
			"count": len(jobs),
		},
	})
}

// Duplicate Detection Endpoints

// ScanDuplicates performs a full duplicate detection scan
// POST /api/admin/duplicate-detection/scan
func (h *AdminHandler) ScanDuplicates(c *gin.Context) {
	var request struct {
		ForceRefresh bool `json:"force_refresh,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		// Allow empty body, use defaults
		request.ForceRefresh = false
	}

	startTime := time.Now()
	report, err := h.embeddingService.ScanForDuplicates(c.Request.Context(), request.ForceRefresh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to scan for duplicates",
			"details": err.Error(),
		})
		return
	}

	report.ProcessingTime = time.Since(startTime).String()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// GetDuplicateResults retrieves stored duplicate detection results
// GET /api/admin/duplicate-detection/results
func (h *AdminHandler) GetDuplicateResults(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	method := c.Query("method") // "embedding", "jaccard", "combined"

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 50
	}

	results, err := h.embeddingService.GetDuplicateResults(limit, method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get duplicate results",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"results": results,
			"count":   len(results),
		},
	})
}

// CheckRecipeDuplicates checks if a specific recipe has duplicates
// POST /api/admin/duplicate-detection/check
func (h *AdminHandler) CheckRecipeDuplicates(c *gin.Context) {
	var recipe models.RecipeData
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid recipe format",
			"details": err.Error(),
		})
		return
	}

	duplicates, err := h.embeddingService.CheckRecipeDuplicates(c.Request.Context(), &recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to check duplicates",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"duplicates_found": len(duplicates) > 0,
			"duplicate_count":  len(duplicates),
			"duplicates":       duplicates,
		},
	})
}

// RefreshEmbedding updates the embedding for a specific recipe
// POST /api/admin/embeddings/refresh/:recipe_id
func (h *AdminHandler) RefreshEmbedding(c *gin.Context) {
	recipeIDStr := c.Param("recipe_id")
	if recipeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Recipe ID is required",
		})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid recipe ID",
		})
		return
	}

	err = h.embeddingService.RefreshEmbedding(c.Request.Context(), recipeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to refresh embedding",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Embedding refreshed successfully",
	})
}

// Metrics Endpoints

// GetTokenUsageMetrics returns current token usage metrics
// GET /api/admin/metrics/token-usage
func (h *AdminHandler) GetTokenUsageMetrics(c *gin.Context) {
	if h.tokenRateLimiter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Token rate limiter not available",
		})
		return
	}

	metrics := h.tokenRateLimiter.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetCostEfficiencyAnalysis returns cost efficiency analysis
// GET /api/admin/metrics/cost-efficiency
func (h *AdminHandler) GetCostEfficiencyAnalysis(c *gin.Context) {
	if h.tokenRateLimiter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Token rate limiter not available",
		})
		return
	}

	costStatus := h.tokenRateLimiter.GetCostStatus()
	metrics := h.tokenRateLimiter.GetMetrics()

	// Calculate efficiency metrics
	efficiency := gin.H{
		"cost_status":   costStatus,
		"token_metrics": metrics,
	}

	if metrics.TotalRequests > 0 {
		efficiency["avg_cost_per_request"] = metrics.EstimatedCostUSD / float64(metrics.TotalRequests)
		efficiency["avg_tokens_per_request"] = float64(metrics.TotalTokensUsed) / float64(metrics.TotalRequests)
	}

	if costStatus.DailyBudgetUSD > 0 {
		efficiency["daily_budget_utilization"] = (costStatus.DailySpentUSD / costStatus.DailyBudgetUSD) * 100
	}

	if costStatus.MonthlyBudgetUSD > 0 {
		efficiency["monthly_budget_utilization"] = (costStatus.MonthlySpentUSD / costStatus.MonthlyBudgetUSD) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    efficiency,
	})
}

// UpdateBudgets updates daily and monthly budgets
// POST /api/admin/metrics/budgets
func (h *AdminHandler) UpdateBudgets(c *gin.Context) {
	if h.tokenRateLimiter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Token rate limiter not available",
		})
		return
	}

	var request struct {
		DailyBudgetUSD   float64 `json:"daily_budget_usd" binding:"min=0"`
		MonthlyBudgetUSD float64 `json:"monthly_budget_usd" binding:"min=0"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	h.tokenRateLimiter.SetBudgets(request.DailyBudgetUSD, request.MonthlyBudgetUSD)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Budgets updated successfully",
		"data": gin.H{
			"daily_budget_usd":   request.DailyBudgetUSD,
			"monthly_budget_usd": request.MonthlyBudgetUSD,
		},
	})
}

// GetSystemHealth returns overall system health status
// GET /api/admin/health
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"services": gin.H{
			"batch_generation":   h.batchService != nil,
			"embedding_service":  h.embeddingService != nil,
			"token_rate_limiter": h.tokenRateLimiter != nil,
			"diversity_service":  h.diversityService != nil,
		},
	}

	// Add metrics if available
	if h.tokenRateLimiter != nil {
		metrics := h.tokenRateLimiter.GetMetrics()
		health["metrics"] = gin.H{
			"total_requests":   metrics.TotalRequests,
			"rate_limit_hits":  metrics.RateLimitHits,
			"requests_blocked": metrics.RequestsBlocked,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

// Diversity System Endpoints (Issue #65)

// GetRecipeCoverage handles GET /api/admin/recipe-coverage
func (h *AdminHandler) GetRecipeCoverage(c *gin.Context) {
	if h.diversityService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Diversity service not available",
		})
		return
	}

	analysis, err := h.diversityService.AnalyzeCoverage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to analyze recipe coverage",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
	})
}

// GenerateDiverseRecipes handles POST /api/admin/generate-diverse
func (h *AdminHandler) GenerateDiverseRecipes(c *gin.Context) {
	if h.diversityService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Diversity service not available",
		})
		return
	}

	var req models.DiverseGenerationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Request validation failed",
			"details": err.Error(),
		})
		return
	}

	// Generate diverse recipes
	response, err := h.diversityService.GenerateDiverseRecipes(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate diverse recipes",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetDiversityMetrics handles GET /api/admin/diversity-metrics
func (h *AdminHandler) GetDiversityMetrics(c *gin.Context) {
	if h.diversityService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Diversity service not available",
		})
		return
	}

	// Get optional limit parameter
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Analyze coverage
	analysis, err := h.diversityService.AnalyzeCoverage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get diversity metrics",
			"details": err.Error(),
		})
		return
	}

	// Limit low coverage combinations if requested
	if len(analysis.LowCoverageCombos) > limit {
		analysis.LowCoverageCombos = analysis.LowCoverageCombos[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
		"meta": gin.H{
			"limit":              limit,
			"total_combos":       analysis.TotalCombinations,
			"covered_combos":     analysis.CoveredCombinations,
			"coverage_rate":      analysis.CoverageRate,
			"low_coverage_count": len(analysis.LowCoverageCombos),
		},
	})
}

// UpdateDimensionWeights handles POST /api/admin/dimension-weights
func (h *AdminHandler) UpdateDimensionWeights(c *gin.Context) {
	var req struct {
		DimensionType  string  `json:"dimension_type" binding:"required"`
		DimensionValue string  `json:"dimension_value" binding:"required"`
		Weight         float64 `json:"weight" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// TODO: Implement dimension weight update in Phase 3
	// This would update the weight in the recipe_dimensions table

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dimension weight update functionality will be implemented in Phase 3",
		"data": gin.H{
			"dimension_type":  req.DimensionType,
			"dimension_value": req.DimensionValue,
			"new_weight":      req.Weight,
		},
	})
}

// InitializeDiversitySystem handles POST /api/admin/initialize-diversity
func (h *AdminHandler) InitializeDiversitySystem(c *gin.Context) {
	if h.diversityService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Diversity service not available",
		})
		return
	}

	if err := h.diversityService.InitializeSystem(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to initialize diversity system",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Diversity system initialized successfully",
	})
}
