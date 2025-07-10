package http

import (
	"net/http"
	"time"

	"enterprise-crud/internal/domain/user"
	userDTO "enterprise-crud/internal/dto/user"
	"enterprise-crud/internal/infrastructure/auth"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations
// Provides REST endpoints for user management
//
// DEPENDENCY INJECTION EXPLANATION:
// - UserHandler depends on user.Service interface (NOT concrete implementation)
// - This follows the Dependency Inversion Principle
// - The handler doesn't know HOW the service works, only WHAT it can do
// - Makes testing easy (can inject mock services)
// - Makes the code flexible (can swap service implementations)
type UserHandler struct {
	userService user.Service // Service layer for user business logic (INTERFACE, not concrete type)
	jwtService  *auth.JWTService // JWT service for token generation and validation
}

// NewUserHandler creates a new instance of UserHandler
//
// DEPENDENCY INJECTION PATTERN:
// - This is a "Constructor" function that receives dependencies
// - userService parameter is an INTERFACE (user.Service)
// - The caller decides which concrete implementation to inject
// - This is "Constructor Injection" - dependencies provided at creation time
//
// WHY THIS PATTERN?
// 1. LOOSE COUPLING: Handler doesn't depend on concrete service implementation
// 2. TESTABILITY: Can inject mock services for testing
// 3. FLEXIBILITY: Can swap service implementations without changing handler code
// 4. SINGLE RESPONSIBILITY: Handler focuses on HTTP concerns, service handles business logic
//
// EXAMPLE USAGE:
// - Production: NewUserHandler(realUserService)
// - Testing: NewUserHandler(mockUserService)
//
// Returns a handler for user HTTP operations
func NewUserHandler(userService user.Service, jwtService *auth.JWTService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// CreateUser handles POST requests to create a new user
// @Summary Create a new user
// @Description Create a new user with email, username and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body userDTO.CreateUserRequest true "User creation request"
// @Success 201 {object} userDTO.UserResponse "User created successfully"
// @Failure 400 {object} userDTO.ErrorResponse "Invalid request data"
// @Failure 409 {object} userDTO.ErrorResponse "User already exists"
// @Failure 500 {object} userDTO.ErrorResponse "Internal server error"
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req userDTO.CreateUserRequest

	// Bind and validate request JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, userDTO.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Call service to create user
	createdUser, err := h.userService.CreateUser(c.Request.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		// Check if error is due to existing user
		if err.Error() == "user with email "+req.Email+" already exists" {
			c.JSON(http.StatusConflict, userDTO.ErrorResponse{
				Error:   "User already exists",
				Message: err.Error(),
			})
			return
		}

		// Handle other creation errors
		c.JSON(http.StatusInternalServerError, userDTO.ErrorResponse{
			Error:   "Failed to create user",
			Message: err.Error(),
		})
		return
	}

	// Return successful response
	response := userDTO.UserResponse{
		ID:       createdUser.ID,
		Email:    createdUser.Email,
		Username: createdUser.Username,
	}

	c.JSON(http.StatusCreated, response)
}

// GetUserByEmail handles GET requests to retrieve a user by email
// @Summary Get user by email
// @Description Get user details by email address
// @Tags users
// @Produce json
// @Param email path string true "User email"
// @Success 200 {object} userDTO.UserResponse "User found"
// @Failure 404 {object} userDTO.ErrorResponse "User not found"
// @Failure 500 {object} userDTO.ErrorResponse "Internal server error"
// @Router /api/v1/users/{email} [get]
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	// Validate email parameter
	if email == "" {
		c.JSON(http.StatusBadRequest, userDTO.ErrorResponse{
			Error:   "Invalid request",
			Message: "Email parameter is required",
		})
		return
	}

	// Call service to get user
	foundUser, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, userDTO.ErrorResponse{
			Error:   "User not found",
			Message: err.Error(),
		})
		return
	}

	// Return successful response
	response := userDTO.UserResponse{
		ID:       foundUser.ID,
		Email:    foundUser.Email,
		Username: foundUser.Username,
	}

	c.JSON(http.StatusOK, response)
}

// Login handles POST requests to authenticate a user
// @Summary User login
// @Description Authenticate user with email and password, returns JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body userDTO.LoginRequest true "User login credentials"
// @Success 200 {object} userDTO.LoginResponse "Login successful"
// @Failure 400 {object} userDTO.ErrorResponse "Invalid request data"
// @Failure 401 {object} userDTO.ErrorResponse "Invalid credentials"
// @Failure 500 {object} userDTO.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req userDTO.LoginRequest

	// Bind and validate request JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, userDTO.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Authenticate user
	authenticatedUser, err := h.userService.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, userDTO.ErrorResponse{
			Error:   "Authentication failed",
			Message: err.Error(),
		})
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(authenticatedUser.ID, authenticatedUser.Email, authenticatedUser.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, userDTO.ErrorResponse{
			Error:   "Failed to generate token",
			Message: err.Error(),
		})
		return
	}

	// Calculate expiration time (matching JWT service expiration)
	expiresAt := time.Now().Add(24 * 30 * time.Hour).Unix() // 30 days (long-lived token)

	// Return successful response
	response := userDTO.LoginResponse{
		User: userDTO.UserResponse{
			ID:       authenticatedUser.ID,
			Email:    authenticatedUser.Email,
			Username: authenticatedUser.Username,
		},
		Token:     token,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile handles GET requests to retrieve the current user's profile
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} userDTO.UserResponse "User profile retrieved successfully"
// @Failure 401 {object} userDTO.ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} userDTO.ErrorResponse "Internal server error"
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user information from JWT token (set by middleware)
	userID, userEmail, userUsername, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, userDTO.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User information not found in token",
		})
		return
	}

	// Return user profile
	response := userDTO.UserResponse{
		ID:       userID,
		Email:    userEmail,
		Username: userUsername,
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers user routes with the gin router
// Sets up POST /users and GET /users/:email endpoints
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Create JWT middleware
	jwtMiddleware := auth.NewJWTMiddleware(h.jwtService)

	// User routes group
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("", h.CreateUser)           // Create new user (public)
		userRoutes.GET("/:email", h.GetUserByEmail) // Get user by email (public)
		
		// Protected routes (require authentication)
		userRoutes.GET("/profile", jwtMiddleware.AuthRequired(), h.GetProfile) // Get current user profile
	}
}

// RegisterAuthRoutes registers authentication routes with the gin router
// Sets up POST /auth/login endpoint
func (h *UserHandler) RegisterAuthRoutes(router *gin.RouterGroup) {
	// Authentication routes group
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", h.Login) // User login
	}
}
