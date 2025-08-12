package middleware

import (
	"net/http"
	"strings"

	"print-service/internal/infrastructure/logger"
	"print-service/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled    bool     `yaml:"enabled"`
	APIKeys    []string `yaml:"api_keys"`
	JWTSecret  string   `yaml:"jwt_secret"`
	RequireSSL bool     `yaml:"require_ssl"`
}

// Auth returns authentication middleware
func Auth(config AuthConfig, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if disabled
		if !config.Enabled {
			c.Next()
			return
		}

		// Require SSL in production
		if config.RequireSSL && c.Request.Header.Get("X-Forwarded-Proto") != "https" && c.Request.TLS == nil {
			logger.Warn("SSL required but request is not secure", "remote_addr", c.ClientIP())
			c.JSON(http.StatusUpgradeRequired, gin.H{
				"error": "SSL required",
			})
			c.Abort()
			return
		}

		// Check for API key in header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Check for API key in Authorization header
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			logger.Warn("Missing API key", "remote_addr", c.ClientIP(), "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
			})
			c.Abort()
			return
		}

		// Validate API key
		if !isValidAPIKey(apiKey, config.APIKeys) {
			logger.Warn("Invalid API key", "remote_addr", c.ClientIP(), "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Set authenticated user context
		c.Set("api_key", apiKey)
		c.Set("authenticated", true)

		logger.Debug("Request authenticated", "remote_addr", c.ClientIP(), "path", c.Request.URL.Path)
		c.Next()
	}
}

// OptionalAuth returns optional authentication middleware for public endpoints
func OptionalAuth(config AuthConfig, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Enabled {
			c.Next()
			return
		}

		// Check for API key but don't require it
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey != "" && isValidAPIKey(apiKey, config.APIKeys) {
			c.Set("api_key", apiKey)
			c.Set("authenticated", true)
			logger.Debug("Optional auth successful", "remote_addr", c.ClientIP())
		} else {
			c.Set("authenticated", false)
		}

		c.Next()
	}
}

// AdminAuth returns admin-only authentication middleware
func AdminAuth(config AuthConfig, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run regular auth
		Auth(config, logger)(c)

		// If auth failed, it would have aborted already
		if c.IsAborted() {
			return
		}

		// Check if this is an admin API key (simplified - in real implementation,
		// you'd have role-based permissions)
		apiKey := c.GetString("api_key")
		if !isAdminAPIKey(apiKey, config.APIKeys) {
			logger.Warn("Admin access denied", "remote_addr", c.ClientIP(), "api_key", maskAPIKey(apiKey))
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Set("admin", true)
		c.Next()
	}
}

// isValidAPIKey checks if the provided API key is valid
func isValidAPIKey(apiKey string, validKeys []string) bool {
	if apiKey == "" {
		return false
	}

	for _, validKey := range validKeys {
		if apiKey == validKey {
			return true
		}
	}
	return false
}

// isAdminAPIKey checks if the API key has admin privileges
// In a real implementation, this would check a database or role system
func isAdminAPIKey(apiKey string, validKeys []string) bool {
	// Simplified: assume first API key in config is admin
	if len(validKeys) > 0 && apiKey == validKeys[0] {
		return true
	}
	return false
}

// maskAPIKey masks an API key for logging (shows only first 4 and last 4 chars)
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

// GenerateAPIKey generates a new API key
func GenerateAPIKey() string {
	return "pk_" + utils.GenerateID()
}
