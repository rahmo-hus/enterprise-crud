package user

import (
	"fmt"
)

// UserError represents domain-specific user errors
type UserError struct {
	Code    string
	Message string
	Cause   error
}

func (e *UserError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *UserError) Unwrap() error {
	return e.Cause
}

// Pre-defined user domain errors
var (
	ErrUserNotFound        = &UserError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrUserAlreadyExists   = &UserError{Code: "USER_EXISTS", Message: "user already exists"}
	ErrInvalidCredentials  = &UserError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}
	ErrPasswordHashFailed  = &UserError{Code: "PASSWORD_HASH_FAILED", Message: "failed to hash password"}
	ErrUserCreationFailed  = &UserError{Code: "USER_CREATION_FAILED", Message: "failed to create user"}
	ErrUserRetrievalFailed = &UserError{Code: "USER_RETRIEVAL_FAILED", Message: "failed to retrieve user"}
	ErrRoleRetrievalFailed = &UserError{Code: "ROLE_RETRIEVAL_FAILED", Message: "failed to retrieve user role"}
)

// NewUserError creates a new UserError with a cause
func NewUserError(baseError *UserError, cause error) *UserError {
	return &UserError{
		Code:    baseError.Code,
		Message: baseError.Message,
		Cause:   cause,
	}
}

// NewUserExistsError creates a specific error for existing users
func NewUserExistsError(email string) *UserError {
	return &UserError{
		Code:    "USER_EXISTS",
		Message: fmt.Sprintf("user with email %s already exists", email),
	}
}
