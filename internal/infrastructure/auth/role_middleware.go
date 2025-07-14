package auth

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// RequireRole creates middleware that checks if the user has any of the allowed roles
// This is like a security guard that checks if you have the right permission to enter
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, make sure the user is authenticated
		// The JWT middleware should have already run and set the user context
		userClaims, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "You must be logged in to access this resource",
			})
			c.Abort()
			return
		}

		// Convert the user context to JWT claims
		claims, ok := userClaims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authentication",
				"message": "Could not verify your authentication credentials",
			})
			c.Abort()
			return
		}

		// Check if the user has any of the required roles
		hasRequiredRole := false
		for _, userRole := range claims.Roles {
			if slices.Contains(allowedRoles, userRole) {
				hasRequiredRole = true
				break
			}
		}

		// If the user doesn't have the required role, deny access
		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Insufficient permissions",
				"message":        "You don't have the required role to access this resource",
				"required_roles": allowedRoles,
				"user_roles":     claims.Roles,
			})
			c.Abort()
			return
		}

		// User has the required role, let them continue
		c.Next()
	}
}

// RequireAdmin is a convenience function that requires ADMIN role
// Use this for endpoints that only administrators should access
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("ADMIN")
}

// RequireUser is a convenience function that requires USER or ADMIN role
// Use this for endpoints that any logged-in user should access
func RequireUser() gin.HandlerFunc {
	return RequireRole("USER", "ADMIN")
}

// RequireOrganizer is a convenience function that requires ORGANIZER or ADMIN role
// Use this for endpoints that only event organizers should access
func RequireOrganizer() gin.HandlerFunc {
	return RequireRole("ORGANIZER", "ADMIN")
}

// GetUserRoles extracts the roles from the current user context
// This helper function can be used in handlers to get the user's roles
func GetUserRoles(c *gin.Context) ([]string, bool) {
	userClaims, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	claims, ok := userClaims.(*JWTClaims)
	if !ok {
		return nil, false
	}

	return claims.Roles, true
}

// HasRole checks if the current user has a specific role
// This helper function can be used in handlers for additional role checks
func HasRole(c *gin.Context, roleName string) bool {
	roles, exists := GetUserRoles(c)
	if !exists {
		return false
	}

	return slices.Contains(roles, roleName)
}
