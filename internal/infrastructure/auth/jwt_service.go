// Package auth provides authentication and authorization services.
// It includes JWT token generation, validation, and role-based access control
// middleware for securing API endpoints.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService handles JWT token operations
type JWTService struct {
	secretKey  []byte
	issuer     string
	expiration time.Duration
}

// JWTClaims represents the JWT claims structure
// Now includes roles for authorization checking
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Roles    []string  `json:"roles"` // Array of role names (ADMIN, USER, etc.)
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service instance
func NewJWTService(secretKey string, issuer string, expiration time.Duration) *JWTService {
	return &JWTService{
		secretKey:  []byte(secretKey),
		issuer:     issuer,
		expiration: expiration,
	}
}

// GenerateToken generates a new JWT token for the user with their roles
func (j *JWTService) GenerateToken(userID uuid.UUID, email, username string, roles []string) (string, error) {
	now := time.Now()

	claims := &JWTClaims{
		UserID:   userID,
		Email:    email,
		Username: username,
		Roles:    roles, // Include user roles in the token
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts the JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}
