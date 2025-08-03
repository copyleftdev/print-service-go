package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that handles CORS headers
func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*") // In production, be more specific
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}
