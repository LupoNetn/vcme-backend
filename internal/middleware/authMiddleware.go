package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/util"
)

// Claims represents the JWT claims we expect in requests.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// AuthMiddleware checks for a valid JWT token in the Authorization header
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		tokenString, err := util.ExtractToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Validate token using local verifier
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
func GetCurrentUser(c *gin.Context) (*Claims, bool) {
	val, ok := c.Get("user")
	if !ok {
		return nil, false
	}

	claims, ok := val.(*Claims)
	//incase error handling
	if !ok {
		return nil, false
	}
	return claims, ok
}
