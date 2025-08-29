package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"lazychef/internal/config"
	"lazychef/internal/database"
	"lazychef/internal/handlers"
	"lazychef/internal/services"
)

func main() {
	// Load environment variables from backend directory
	if err := godotenv.Load("../../.env"); err != nil {
		// Try loading from backend directory as fallback
		if err2 := godotenv.Load("../.env"); err2 != nil {
			log.Println("Warning: .env file not found, using environment variables")
		}
	}

	// Initialize database
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/recipes.db"
	}

	dbConfig := database.Config{
		Path: dbPath,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Load OpenAI configuration
	openaiConfig, err := config.LoadOpenAIConfig()
	if err != nil {
		log.Printf("Warning: OpenAI configuration error: %v", err)
		log.Println("Recipe generation will not be available")
	}

	// Initialize services
	var recipeHandler *handlers.RecipeHandler
	var mealPlanHandler *handlers.MealPlanHandler
	var adminHandler *handlers.AdminHandler

	if openaiConfig != nil {
		// Initialize legacy generator service
		generatorService, err := services.NewRecipeGeneratorService(openaiConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize recipe generator: %v", err)
		} else {
			// Initialize enhanced generator service
			enhancedGeneratorService := services.NewEnhancedRecipeGeneratorService(
				generatorService.GetClient(),
				openaiConfig,
				generatorService.GetRateLimiter(),
				generatorService.GetCache(),
			)

			recipeHandler = handlers.NewRecipeHandler(db, generatorService, enhancedGeneratorService)

			// Initialize meal planner with database and generator
			mealPlannerService := services.NewMealPlannerService(db, generatorService)
			mealPlanHandler = handlers.NewMealPlanHandler(mealPlannerService)

			// Initialize Phase 1 services
			batchStoragePath := os.Getenv("BATCH_STORAGE_PATH")
			if batchStoragePath == "" {
				batchStoragePath = "./data/batch_files"
			}

			// Batch generation service
			batchService := services.NewBatchGenerationService(
				generatorService.GetClient(),
				openaiConfig,
				db.DB,
				batchStoragePath,
			)

			// Embedding deduplicator
			embeddingService := services.NewEmbeddingDeduplicator(
				generatorService.GetClient(),
				db.DB,
			)

			// Advanced token rate limiter
			tokenRateLimiter := services.NewTokenRateLimiter(
				openaiConfig.RequestsPerMinute,
				1000,   // tokens per second
				100.0,  // daily budget USD
				3000.0, // monthly budget USD
			)

			// Admin handler for new APIs
			adminHandler = handlers.NewAdminHandler(
				batchService,
				embeddingService,
				tokenRateLimiter,
			)

			log.Printf("GPT-5 Enhanced Services Initialized:")
			log.Printf("  - Ideation Model: %s", openaiConfig.IdeationModel)
			log.Printf("  - Authoring Model: %s", openaiConfig.AuthoringModel)
			log.Printf("  - Critique Model: %s", openaiConfig.CritiqueModel)
			log.Printf("  - Structured Outputs: %t", openaiConfig.UseStructuredOutputs)
			log.Printf("  - Food Safety Strict Mode: %t", openaiConfig.FoodSafetyStrictMode)

			log.Printf("Phase 1 Services Initialized:")
			log.Printf("  - Batch API Service: enabled")
			log.Printf("  - Embedding Deduplicator: enabled")
			log.Printf("  - Token Rate Limiter: enabled")
			log.Printf("  - Batch Storage Path: %s", batchStoragePath)
		}
	}

	// Setup Gin router
	r := gin.Default()

	// CORS middleware - permissive for local development
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		health := gin.H{
			"status":            "ok",
			"service":           "lazychef-api",
			"openai_configured": openaiConfig != nil,
		}

		if openaiConfig != nil {
			health["gpt5_features"] = gin.H{
				"structured_outputs": openaiConfig.UseStructuredOutputs,
				"food_safety_mode":   openaiConfig.FoodSafetyStrictMode,
				"models": gin.H{
					"ideation":  openaiConfig.IdeationModel,
					"authoring": openaiConfig.AuthoringModel,
					"critique":  openaiConfig.CritiqueModel,
				},
			}
		}

		c.JSON(200, health)
	})

	// Recipe generation endpoints (only if OpenAI is configured)
	if recipeHandler != nil {
		api := r.Group("/api/recipes")
		{
			// Legacy endpoints
			api.POST("/generate", recipeHandler.GenerateRecipe)
			api.POST("/generate-batch", recipeHandler.GenerateBatchRecipes)
			api.GET("/health", recipeHandler.GetGeneratorHealth)
			api.POST("/clear-cache", recipeHandler.ClearCache)
			api.GET("/test", recipeHandler.TestRecipeGeneration)
			api.GET("/search", recipeHandler.SearchRecipes)

			// Enhanced GPT-5 endpoints
			api.POST("/generate-enhanced", recipeHandler.GenerateRecipeEnhanced)
			api.POST("/validate-safety", recipeHandler.ValidateRecipeSafety)
			api.POST("/validate-quality", recipeHandler.ValidateRecipeQuality)
		}
	} else {
		// Fallback endpoints when OpenAI is not configured
		r.GET("/api/recipes/test", func(c *gin.Context) {
			c.JSON(503, gin.H{
				"error":   "OpenAI API not configured",
				"message": "Please set OPENAI_API_KEY environment variable",
			})
		})

		// Basic search endpoint even without OpenAI
		r.GET("/api/recipes/search", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"data": gin.H{
					"recipes": []interface{}{},
					"total":   0,
				},
			})
		})
	}

	// Meal planning endpoints
	if mealPlanHandler != nil {
		mealPlanAPI := r.Group("/api/meal-plans")
		{
			mealPlanAPI.POST("/create", mealPlanHandler.CreateMealPlan)
			mealPlanAPI.GET("/:id", mealPlanHandler.GetMealPlan)
			mealPlanAPI.GET("/", mealPlanHandler.ListMealPlans)
		}
	} else {
		// Fallback meal plan endpoints
		r.POST("/api/meal-plans/create", func(c *gin.Context) {
			c.JSON(503, gin.H{
				"error":   "Meal planning service not available",
				"message": "Please configure OpenAI API to enable meal planning",
			})
		})
	}

	// Admin endpoints for Phase 1 features
	if adminHandler != nil {
		adminAPI := r.Group("/api/admin")
		{
			// Batch generation endpoints
			batchAPI := adminAPI.Group("/batch-generation")
			{
				batchAPI.POST("/submit", adminHandler.SubmitBatchGeneration)
				batchAPI.GET("/status/:job_id", adminHandler.GetBatchStatus)
				batchAPI.POST("/cancel/:job_id", adminHandler.CancelBatchJob)
				batchAPI.GET("/results/:job_id", adminHandler.GetBatchResults)
				batchAPI.GET("/jobs", adminHandler.ListBatchJobs)
			}

			// Duplicate detection endpoints
			duplicateAPI := adminAPI.Group("/duplicate-detection")
			{
				duplicateAPI.POST("/scan", adminHandler.ScanDuplicates)
				duplicateAPI.GET("/results", adminHandler.GetDuplicateResults)
				duplicateAPI.POST("/check", adminHandler.CheckRecipeDuplicates)
			}

			// Embedding endpoints
			embeddingAPI := adminAPI.Group("/embeddings")
			{
				embeddingAPI.POST("/refresh/:recipe_id", adminHandler.RefreshEmbedding)
			}

			// Metrics endpoints
			metricsAPI := adminAPI.Group("/metrics")
			{
				metricsAPI.GET("/token-usage", adminHandler.GetTokenUsageMetrics)
				metricsAPI.GET("/cost-efficiency", adminHandler.GetCostEfficiencyAnalysis)
				metricsAPI.POST("/budgets", adminHandler.UpdateBudgets)
			}

			// System health
			adminAPI.GET("/health", adminHandler.GetSystemHealth)
		}
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("LazyChef API server starting on port %s", port)
	log.Printf("OpenAI configured: %t", openaiConfig != nil)
	log.Printf("Health check: http://localhost:%s/api/health", port)

	if recipeHandler != nil {
		log.Printf("Recipe test: http://localhost:%s/api/recipes/test", port)
		log.Printf("Enhanced generation: http://localhost:%s/api/recipes/generate-enhanced", port)
		log.Printf("Safety validation: http://localhost:%s/api/recipes/validate-safety", port)
		log.Printf("Quality validation: http://localhost:%s/api/recipes/validate-quality", port)
	}

	if adminHandler != nil {
		log.Printf("Admin endpoints available:")
		log.Printf("  - Batch jobs: http://localhost:%s/api/admin/batch-generation/jobs", port)
		log.Printf("  - Duplicate scan: http://localhost:%s/api/admin/duplicate-detection/scan", port)
		log.Printf("  - Token metrics: http://localhost:%s/api/admin/metrics/token-usage", port)
		log.Printf("  - Admin health: http://localhost:%s/api/admin/health", port)
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
