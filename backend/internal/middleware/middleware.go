// Package middleware provides HTTP middleware for the LazyChef API
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SetupMiddleware configures all middleware for the Gin router
func SetupMiddleware(r *gin.Engine) {
	// Request ID middleware (first to ensure all requests have ID)
	r.Use(RequestIDMiddleware())
	
	// Error handling middleware (early to catch all errors)
	r.Use(ErrorHandlerMiddleware())
	
	// Logging middleware
	r.Use(LoggerMiddleware())
	
	// CORS middleware
	r.Use(CORSMiddleware())
}