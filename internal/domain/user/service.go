package user

import (
	"context"
	"enterpr
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service defines the business logic interface for user operations
// This is similar to Spring Boot's @Service layer - handles business rules and validations
// Orchestrates between the repository layer and the presentation layer
type Service interface {
	CreateUser(ctx context.Context, email, username, password string) (*User, error) // Creates a new user with validation and password hashing
	GetUserByEmail(ctx context.Context, email string) (*User, error)                 // Retrieves a user by email with business logic
	AuthenticateUser(ctx context.Context, email, password string) (*User, error)     // Authenticates user with email and password
}

// userService implements the Service interface
// This is the concrete implementation of business logic, similar to Spring Boot's @Service classes
// Encapsulates all user-related business operations and rules
type userService struct {
	repo     Repository      // Repository dependency for data persistence - similar to @Autowired in Spring
	roleRepo role.Repository // Role repository to assign default roles to users
}

// NewUserService creates a new instance of userService
// Returns a service implementation for user business logic
func NewUserService(repo Repository, roleRepo role.Repository) Service {
	return &userService{
		repo:     repo,
		roleRepo: roleRepo,
	}
}

// CreateUser creates a new user with the provided information
//
// BUSINESS LOGIC FLOW:
// 1. Validation: Check if user already exists (business rule)
// 2. Security: Hash password before storage (security requirement)
// 3. Entity Creation: Create domain entity with all required fields
// 4. Persistence: Save to database via repository
// 5. Error Handling: Wrap and contextualize any errors
//
// WHY THIS PATTERN?
// - Encapsulates business rules in one place
// - Handles complex operations that span multiple steps
// - Provides transaction-like behavior
// - Separates business logic from data access
// - Makes testing easier (can mock repository)
//
// Validates input, hashes password, and persists user data
func (s *userService) CreateUser(ctx context.Context, email, username, password string) (*User, error) {
	// STEP 1: BUSINESS RULE VALIDATION
	// Check if user already exists with this email
	// This is a business rule: "Users must have unique email addresses"
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// Return business logic error (not a system error)
		return nil, NewUserExistsError(email)
	}
	// If error is not "record not found", it's a database error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NewUserError(ErrUserRetrievalFailed, err)
	}

	// STEP 2: SECURITY IMPLEMENTATION
	// Hash the password for secure storage
	// bcrypt.DefaultCost provides good security vs. performance balance
	// Never store plain text passwords in database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// Return wrapped error with context
		return nil, NewUserError(ErrPasswordHashFailed, err)
	}

	// STEP 3: GET DEFAULT USER ROLE
	// Every new user gets the "USER" role by default
	// This is a business rule: all registered users start as regular users
	userRole, err := s.roleRepo.GetByName(ctx, role.RoleUser)
	if err != nil {
		return nil, NewUserError(ErrRoleRetrievalFailed, err)
	}

	// STEP 4: DOMAIN ENTITY CREATION
	// Create new user entity with all required fields and default role
	user := &User{
		ID:       uuid.New(),             // Generate unique identifier (UUID v4)
		Email:    email,                  // Set email address (validated by HTTP layer)
		Username: username,               // Set username (validated by HTTP layer)
		Password: string(hashedPassword), // Store hashed password
		Roles:    []role.Role{*userRole}, // Assign default USER role
	}

	// STEP 5: PERSIST USER TO DATABASE
	// This will save both the user and the role assignment
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, NewUserError(ErrUserCreationFailed, err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email address
// Returns user if found, error if not found or database error occurs
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, NewUserError(ErrUserRetrievalFailed, err)
	}
	return user, nil
}

// AuthenticateUser validates user credentials and returns user if valid
//
// AUTHENTICATION FLOW:
// 1. Retrieve user by email from database
// 2. Compare provided password with stored hashed password
// 3. Return user if passwords match, error if not
//
// SECURITY CONSIDERATIONS:
// - Uses bcrypt for password verification (secure against timing attacks)
// - Never returns the hashed password to prevent exposure
// - Provides generic error messages to prevent user enumeration
func (s *userService) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	// STEP 1: GET USER BY EMAIL
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// Return generic error to prevent user enumeration
		return nil, ErrInvalidCredentials
	}

	// STEP 2: VERIFY PASSWORD
	// bcrypt.CompareHashAndPassword is secure against timing attacks
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Return generic error to prevent user enumeration
		return nil, ErrInvalidCredentials
	}

	// STEP 3: RETURN AUTHENTICATED USER
	// Password verification successful
	return user, nil
}
