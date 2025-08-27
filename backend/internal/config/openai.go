package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey              string
	Model               string
	MaxTokens           int
	Temperature         float32
	RequestTimeout      time.Duration
	MaxRetries          int
	RetryDelay          time.Duration
	RequestsPerMinute   int
	RecipeGenerationTimeout time.Duration
}

// LoadOpenAIConfig loads OpenAI configuration from environment variables
func LoadOpenAIConfig() (*OpenAIConfig, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is required")
	}
	
	config := &OpenAIConfig{
		APIKey:      apiKey,
		Model:       getEnvOrDefault("OPENAI_MODEL", "gpt-3.5-turbo"),
		MaxTokens:   getEnvAsIntOrDefault("OPENAI_MAX_TOKENS", 1000),
		Temperature: getEnvAsFloatOrDefault("OPENAI_TEMPERATURE", 0.7),
		RequestTimeout: getEnvAsDurationOrDefault("OPENAI_REQUEST_TIMEOUT", 30*time.Second),
		MaxRetries:     getEnvAsIntOrDefault("OPENAI_MAX_RETRIES", 3),
		RetryDelay:     getEnvAsDurationOrDefault("OPENAI_RETRY_DELAY", 2*time.Second),
		RequestsPerMinute: getEnvAsIntOrDefault("OPENAI_REQUESTS_PER_MINUTE", 60),
		RecipeGenerationTimeout: getEnvAsDurationOrDefault("RECIPE_GENERATION_TIMEOUT", 30*time.Second),
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

// Validate validates the OpenAI configuration
func (c *OpenAIConfig) Validate() error {
	if c.APIKey == "" {
		return errors.New("API key cannot be empty")
	}
	if c.MaxTokens <= 0 {
		return errors.New("max tokens must be positive")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return errors.New("temperature must be between 0 and 2")
	}
	if c.RequestTimeout <= 0 {
		return errors.New("request timeout must be positive")
	}
	if c.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}
	if c.RequestsPerMinute <= 0 {
		return errors.New("requests per minute must be positive")
	}
	return nil
}

// GetRateLimitDelay calculates delay between requests based on rate limit
func (c *OpenAIConfig) GetRateLimitDelay() time.Duration {
	return time.Minute / time.Duration(c.RequestsPerMinute)
}

// IsProduction checks if we're using production OpenAI model
func (c *OpenAIConfig) IsProduction() bool {
	return c.Model == "gpt-4" || c.Model == "gpt-4-turbo"
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloatOrDefault(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}