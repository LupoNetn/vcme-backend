package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/luponetn/vcme/internal/config"
)

// Claims represents the JWT claims we expect in requests.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// verifyToken parses and validates the JWT using the provided secret.
func verifyToken(tokenString, secret string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("empty token")
	}
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

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

		// Validate token using local verifier
		claims, err := verifyToken(tokenString, cfg.JWTAccessSecret)
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
	return claims, ok
}
