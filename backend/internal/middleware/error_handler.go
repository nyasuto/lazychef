package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if _, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "Internal Server Error",
				"message":    "Something went wrong",
				"request_id": getRequestID(c),
			})
			c.Abort()
			return
		}

		// For other types of recovered values
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal Server Error",
			"message":    "An unexpected error occurred",
			"request_id": getRequestID(c),
		})
		c.Abort()
	})
}

// getRequestID extracts request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return "unknown"
}