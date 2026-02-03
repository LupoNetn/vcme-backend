package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/util"
)

// AuthMiddleware checks for a valid JWT token in the Authorization header
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header must start with Bearer"})
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := util.VerifyToken(tokenString, cfg.JWTAccessSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", claims)

		c.Next()
	}
}

// GetCurrentUser extracts the user from the context
func GetCurrentUser(c *gin.Context) (*util.Claims, bool) {
	val, ok := c.Get("user")
	if !ok {
		return nil, false
	}

	claims, ok := val.(*util.Claims)
	return claims, ok
}
