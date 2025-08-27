package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOpenAIConfig_WithValidEnv(t *testing.T) {
	// Set up environment variables
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalModel := os.Getenv("OPENAI_MODEL")

	defer func() {
		// Restore original values
		_ = os.Setenv("OPENAI_API_KEY", originalAPIKey)
		_ = os.Setenv("OPENAI_MODEL", originalModel)
	}()

	_ = os.Setenv("OPENAI_API_KEY", "test-api-key")
	_ = os.Setenv("OPENAI_MODEL", "gpt-4")

	config, err := LoadOpenAIConfig()

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-api-key", config.APIKey)
	assert.Equal(t, "gpt-4", config.Model)
}

func TestLoadOpenAIConfig_WithDefaults(t *testing.T) {
	// Set up environment variables with only required key
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalModel := os.Getenv("OPENAI_MODEL")
	originalTemp := os.Getenv("OPENAI_TEMPERATURE")

	defer func() {
		// Restore original values
		_ = os.Setenv("OPENAI_API_KEY", originalAPIKey)
		_ = os.Setenv("OPENAI_MODEL", originalModel)
		_ = os.Setenv("OPENAI_TEMPERATURE", originalTemp)
	}()

	_ = os.Setenv("OPENAI_API_KEY", "test-api-key")
	_ = os.Unsetenv("OPENAI_MODEL")
	_ = os.Unsetenv("OPENAI_TEMPERATURE")

	config, err := LoadOpenAIConfig()

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-api-key", config.APIKey)
	assert.Equal(t, "gpt-3.5-turbo", config.Model)         // Default
	assert.Equal(t, 0.7, config.Temperature)               // Default
	assert.Equal(t, 1000, config.MaxTokens)                // Default
	assert.Equal(t, 60, config.RequestsPerMinute)          // Default
	assert.Equal(t, 3, config.MaxRetries)                  // Default
	assert.Equal(t, time.Second, config.RetryDelay)        // Default
	assert.Equal(t, 30*time.Second, config.RequestTimeout) // Default
}

func TestLoadOpenAIConfig_MissingAPIKey(t *testing.T) {
	// Save original API key
	originalAPIKey := os.Getenv("OPENAI_API_KEY")

	defer func() {
		// Restore original value
		_ = os.Setenv("OPENAI_API_KEY", originalAPIKey)
	}()

	_ = os.Unsetenv("OPENAI_API_KEY")

	config, err := LoadOpenAIConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY")
}

func TestLoadOpenAIConfig_CustomValues(t *testing.T) {
	// Set up custom environment variables
	originalValues := map[string]string{
		"OPENAI_API_KEY":             os.Getenv("OPENAI_API_KEY"),
		"OPENAI_MODEL":               os.Getenv("OPENAI_MODEL"),
		"OPENAI_TEMPERATURE":         os.Getenv("OPENAI_TEMPERATURE"),
		"OPENAI_MAX_TOKENS":          os.Getenv("OPENAI_MAX_TOKENS"),
		"OPENAI_REQUESTS_PER_MINUTE": os.Getenv("OPENAI_REQUESTS_PER_MINUTE"),
		"OPENAI_MAX_RETRIES":         os.Getenv("OPENAI_MAX_RETRIES"),
		"OPENAI_RETRY_DELAY":         os.Getenv("OPENAI_RETRY_DELAY"),
		"OPENAI_REQUEST_TIMEOUT":     os.Getenv("OPENAI_REQUEST_TIMEOUT"),
	}

	defer func() {
		// Restore original values
		for key, value := range originalValues {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	_ = os.Setenv("OPENAI_API_KEY", "custom-key")
	_ = os.Setenv("OPENAI_MODEL", "gpt-4")
	_ = os.Setenv("OPENAI_TEMPERATURE", "0.5")
	_ = os.Setenv("OPENAI_MAX_TOKENS", "2000")
	_ = os.Setenv("OPENAI_REQUESTS_PER_MINUTE", "120")
	_ = os.Setenv("OPENAI_MAX_RETRIES", "5")
	_ = os.Setenv("OPENAI_RETRY_DELAY", "2s")
	_ = os.Setenv("OPENAI_REQUEST_TIMEOUT", "60s")

	config, err := LoadOpenAIConfig()

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "custom-key", config.APIKey)
	assert.Equal(t, "gpt-4", config.Model)
	assert.Equal(t, 0.5, config.Temperature)
	assert.Equal(t, 2000, config.MaxTokens)
	assert.Equal(t, 120, config.RequestsPerMinute)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryDelay)
	assert.Equal(t, 60*time.Second, config.RequestTimeout)
}

func TestLoadOpenAIConfig_InvalidTemperature(t *testing.T) {
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalTemp := os.Getenv("OPENAI_TEMPERATURE")

	defer func() {
		_ = os.Setenv("OPENAI_API_KEY", originalAPIKey)
		_ = os.Setenv("OPENAI_TEMPERATURE", originalTemp)
	}()

	_ = os.Setenv("OPENAI_API_KEY", "test-key")
	_ = os.Setenv("OPENAI_TEMPERATURE", "invalid")

	config, err := LoadOpenAIConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "temperature")
}

func TestLoadOpenAIConfig_InvalidMaxTokens(t *testing.T) {
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalMaxTokens := os.Getenv("OPENAI_MAX_TOKENS")

	defer func() {
		_ = os.Setenv("OPENAI_API_KEY", originalAPIKey)
		_ = os.Setenv("OPENAI_MAX_TOKENS", originalMaxTokens)
	}()

	_ = os.Setenv("OPENAI_API_KEY", "test-key")
	_ = os.Setenv("OPENAI_MAX_TOKENS", "not-a-number")

	config, err := LoadOpenAIConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "max_tokens")
}

func TestLoadOpenAIConfig_InvalidDuration(t *testing.T) {
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalTimeout := os.Getenv("OPENAI_REQUEST_TIMEOUT")

	defer func() {
		_ = os.Setenv("OPENAI_API_KEY", originalAPIKey)
		_ = os.Setenv("OPENAI_REQUEST_TIMEOUT", originalTimeout)
	}()

	_ = os.Setenv("OPENAI_API_KEY", "test-key")
	_ = os.Setenv("OPENAI_REQUEST_TIMEOUT", "invalid-duration")

	config, err := LoadOpenAIConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "request_timeout")
}
