package models

import "errors"

// Recipe validation errors
var (
	ErrInvalidTitle         = errors.New("recipe title cannot be empty")
	ErrInvalidCookingTime   = errors.New("cooking time must be greater than 0")
	ErrNoIngredients        = errors.New("recipe must have at least one ingredient")
	ErrNoSteps              = errors.New("recipe must have at least one cooking step")
	ErrInvalidLazinessScore = errors.New("laziness score must be between 1.0 and 10.0")
)

// Meal plan validation errors
var (
	ErrInvalidStartDate    = errors.New("start date cannot be empty")
	ErrEmptyShoppingList   = errors.New("shopping list cannot be empty")
	ErrNoDailyRecipes      = errors.New("meal plan must have daily recipes")
	ErrInsufficientRecipes = errors.New("meal plan must have at least 3 days of recipes")
)

// Database errors
var (
	ErrRecipeNotFound     = errors.New("recipe not found")
	ErrMealPlanNotFound   = errors.New("meal plan not found")
	ErrUserNotFound       = errors.New("user preferences not found")
	ErrDatabaseConnection = errors.New("failed to connect to database")
	ErrInvalidJSON        = errors.New("invalid JSON data")
)

// API errors
var (
	ErrInvalidRequest    = errors.New("invalid request format")
	ErrMissingParameters = errors.New("missing required parameters")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// OpenAI service errors
var (
	ErrOpenAIConnection = errors.New("failed to connect to OpenAI API")
	ErrOpenAIRateLimit  = errors.New("OpenAI API rate limit exceeded")
	ErrOpenAIInvalidKey = errors.New("invalid OpenAI API key")
	ErrRecipeGeneration = errors.New("failed to generate recipe")
)

// Validation errors
var (
	ErrInvalidSeason     = errors.New("invalid season, must be spring, summer, fall, winter, or all")
	ErrInvalidDifficulty = errors.New("invalid difficulty, must be easy, medium, or hard")
	ErrInvalidSkillLevel = errors.New("invalid skill level, must be beginner, intermediate, or advanced")
)

// Diversity system errors
var (
	ErrInvalidProfileName      = errors.New("invalid profile name")
	ErrInvalidBatchSize        = errors.New("invalid batch size, must be between 1 and 50")
	ErrInvalidStrategy         = errors.New("invalid generation strategy")
	ErrInvalidSimilarity       = errors.New("invalid similarity threshold, must be between 0.0 and 1.0")
	ErrInvalidQualityThreshold = errors.New("invalid quality threshold, must be between 1.0 and 10.0")
	ErrInvalidDimension        = errors.New("invalid dimension")
	ErrDimensionNotFound       = errors.New("dimension not found")
	ErrProfileNotFound         = errors.New("generation profile not found")
	ErrInvalidWeight           = errors.New("invalid dimension weight")
	ErrCoverageNotFound        = errors.New("coverage data not found")
)
