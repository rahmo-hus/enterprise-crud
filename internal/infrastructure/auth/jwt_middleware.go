package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// JWTMiddleware provides JWT authentication middleware
type JWTMiddleware struct {
	jwtService *JWTService
}

// NewJWTMiddleware creates a new JWT middleware instance
func NewJWTMiddleware(jwtService *JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
	}
}

// AuthRequired middleware that requires valid JWT token
func (m *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization required",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// OptionalAuth middleware that optionally validates JWT token
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.Next()
			return
		}

		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.Next()
			return
		}

		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// GetUserFromContext extracts user information from Gin context
func GetUserFromContext(c *gin.Context) (uuid.UUID, string, string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, "", "", false
	}

	email, exists := c.Get("user_email")
	if !exists {
		return uuid.Nil, "", "", false
	}

	username, exists := c.Get("user_username")
	if !exists {
		return uuid.Nil, "", "", false
	}

	return userID.(uuid.UUID), email.(string), username.(string), true
}

// GetClaimsFromContext extracts JWT claims from Gin context
func GetClaimsFromContext(c *gin.Context) (*JWTClaims, bool) {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*JWTClaims)
	return jwtClaims, ok
}
