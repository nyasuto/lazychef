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

	if openaiConfig != nil {
		generatorService, err := services.NewRecipeGeneratorService(openaiConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize recipe generator: %v", err)
		} else {
			recipeHandler = handlers.NewRecipeHandler(generatorService)

			// Initialize meal planner with database and generator
			mealPlannerService := services.NewMealPlannerService(db, generatorService)
			mealPlanHandler = handlers.NewMealPlanHandler(mealPlannerService)
		}
	}

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":            "ok",
			"service":           "lazychef-api",
			"openai_configured": openaiConfig != nil,
		})
	})

	// Recipe generation endpoints (only if OpenAI is configured)
	if recipeHandler != nil {
		api := r.Group("/api/recipes")
		{
			api.POST("/generate", recipeHandler.GenerateRecipe)
			api.POST("/generate-batch", recipeHandler.GenerateBatchRecipes)
			api.GET("/health", recipeHandler.GetGeneratorHealth)
			api.POST("/clear-cache", recipeHandler.ClearCache)
			api.GET("/test", recipeHandler.TestRecipeGeneration)
			api.GET("/search", recipeHandler.SearchRecipes)
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
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
