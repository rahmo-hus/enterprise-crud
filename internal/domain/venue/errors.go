package venue

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// VenueError represents domain-specific venue errors
type VenueError struct {
	Code    string
	Message string
	Cause   error
}

func (e *VenueError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *VenueError) Unwrap() error {
	return e.Cause
}

// Pre-defined venue domain errors
var (
	ErrVenueNotFound        = &VenueError{Code: "VENUE_NOT_FOUND", Message: "venue not found"}
	ErrVenueAlreadyExists   = &VenueError{Code: "VENUE_EXISTS", Message: "venue already exists"}
	ErrVenueCreationFailed  = &VenueError{Code: "VENUE_CREATION_FAILED", Message: "failed to create venue"}
	ErrVenueUpdateFailed    = &VenueError{Code: "VENUE_UPDATE_FAILED", Message: "failed to update venue"}
	ErrVenueDeletionFailed  = &VenueError{Code: "VENUE_DELETION_FAILED", Message: "failed to delete venue"}
	ErrVenueRetrievalFailed = &VenueError{Code: "VENUE_RETRIEVAL_FAILED", Message: "failed to retrieve venue"}
	ErrInvalidVenueCapacity = &VenueError{Code: "INVALID_VENUE_CAPACITY", Message: "venue capacity must be greater than 0"}
)

// NewVenueError creates a new VenueError with a cause
func NewVenueError(baseError *VenueError, cause error) *VenueError {
	return &VenueError{
		Code:    baseError.Code,
		Message: baseError.Message,
		Cause:   cause,
	}
}

// NewVenueNotFoundError creates a specific error for venue not found
func NewVenueNotFoundError(venueID uuid.UUID) *VenueError {
	return &VenueError{
		Code:    "VENUE_NOT_FOUND",
		Message: fmt.Sprintf("venue with ID %s not found", venueID),
	}
}

// IsVenueError checks if an error is a VenueError
func IsVenueError(err error) bool {
	var venueErr *VenueError
	return errors.As(err, &venueErr)
}

// GetVenueErrorCode extracts the error code from a VenueError
func GetVenueErrorCode(err error) string {
	var venueErr *VenueError
	if errors.As(err, &venueErr) {
		return venueErr.Code
	}
	return ""
}

// IsVenueNotFoundError checks if an error is a "not found" error
func IsVenueNotFoundError(err error) bool {
	var venueErr *VenueError
	return errors.As(err, &venueErr) && venueErr.Code == "VENUE_NOT_FOUND"
}
