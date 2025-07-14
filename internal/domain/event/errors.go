package event

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// EventError represents domain-specific event errors
type EventError struct {
	Code    string
	Message string
	Cause   error
}

func (e *EventError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *EventError) Unwrap() error {
	return e.Cause
}

// Pre-defined event domain errors
var (
	ErrEventNotFound           = &EventError{Code: "EVENT_NOT_FOUND", Message: "event not found"}
	ErrEventAlreadyExists      = &EventError{Code: "EVENT_EXISTS", Message: "event already exists"}
	ErrEventCreationFailed     = &EventError{Code: "EVENT_CREATION_FAILED", Message: "failed to create event"}
	ErrEventUpdateFailed       = &EventError{Code: "EVENT_UPDATE_FAILED", Message: "failed to update event"}
	ErrEventDeletionFailed     = &EventError{Code: "EVENT_DELETION_FAILED", Message: "failed to delete event"}
	ErrEventRetrievalFailed    = &EventError{Code: "EVENT_RETRIEVAL_FAILED", Message: "failed to retrieve event"}
	ErrVenueNotFound           = &EventError{Code: "VENUE_NOT_FOUND", Message: "venue not found"}
	ErrEventDateInPast         = &EventError{Code: "EVENT_DATE_INVALID", Message: "event date must be in the future"}
	ErrTicketsExceedCapacity   = &EventError{Code: "TICKETS_EXCEED_CAPACITY", Message: "total tickets exceed venue capacity"}
	ErrUnauthorizedAccess      = &EventError{Code: "UNAUTHORIZED_ACCESS", Message: "only event organizer can perform this action"}
	ErrEventAlreadyCancelled   = &EventError{Code: "EVENT_ALREADY_CANCELLED", Message: "event is already cancelled"}
	ErrEventAlreadyCompleted   = &EventError{Code: "EVENT_ALREADY_COMPLETED", Message: "event is already completed"}
	ErrCannotCancelCompleted   = &EventError{Code: "CANNOT_CANCEL_COMPLETED", Message: "cannot cancel a completed event"}
	ErrCannotUpdateCancelled   = &EventError{Code: "CANNOT_UPDATE_CANCELLED", Message: "cannot update cancelled event"}
	ErrCannotUpdateCompleted   = &EventError{Code: "CANNOT_UPDATE_COMPLETED", Message: "cannot update completed event"}
	ErrCannotDeleteWithTickets = &EventError{Code: "CANNOT_DELETE_WITH_TICKETS", Message: "cannot delete event with sold tickets"}
	ErrInvalidTicketReduction  = &EventError{Code: "INVALID_TICKET_REDUCTION", Message: "cannot reduce total tickets below sold tickets"}
)

// NewEventError creates a new EventError with a cause
func NewEventError(baseError *EventError, cause error) *EventError {
	return &EventError{
		Code:    baseError.Code,
		Message: baseError.Message,
		Cause:   cause,
	}
}

// NewEventNotFoundError creates a specific error for event not found
func NewEventNotFoundError(eventID uuid.UUID) *EventError {
	return &EventError{
		Code:    "EVENT_NOT_FOUND",
		Message: fmt.Sprintf("event with ID %s not found", eventID),
	}
}

// NewVenueNotFoundError creates a specific error for venue not found
func NewVenueNotFoundError(venueID uuid.UUID) *EventError {
	return &EventError{
		Code:    "VENUE_NOT_FOUND",
		Message: fmt.Sprintf("venue with ID %s not found", venueID),
	}
}

// NewTicketsExceedCapacityError creates a specific error for ticket capacity validation
func NewTicketsExceedCapacityError(totalTickets, venueCapacity int) *EventError {
	return &EventError{
		Code:    "TICKETS_EXCEED_CAPACITY",
		Message: fmt.Sprintf("total tickets (%d) cannot exceed venue capacity (%d)", totalTickets, venueCapacity),
	}
}

// NewInvalidTicketReductionError creates a specific error for invalid ticket reduction
func NewInvalidTicketReductionError(requestedTotal, soldTickets int) *EventError {
	return &EventError{
		Code:    "INVALID_TICKET_REDUCTION",
		Message: fmt.Sprintf("cannot reduce total tickets to %d, %d tickets already sold", requestedTotal, soldTickets),
	}
}

// NewUnauthorizedAccessError creates a specific error for unauthorized access
func NewUnauthorizedAccessError(action string) *EventError {
	return &EventError{
		Code:    "UNAUTHORIZED_ACCESS",
		Message: fmt.Sprintf("only event organizer can %s", action),
	}
}

// IsEventNotFoundError checks if an error is a "not found" error
func IsEventNotFoundError(err error) bool {
	var eventErr *EventError
	return errors.As(err, &eventErr) && eventErr.Code == "EVENT_NOT_FOUND"
}

// IsVenueNotFoundError checks if an error is a "venue not found" error
func IsVenueNotFoundError(err error) bool {
	var eventErr *EventError
	return errors.As(err, &eventErr) && eventErr.Code == "VENUE_NOT_FOUND"
}

// IsUnauthorizedError checks if an error is an unauthorized access error
func IsUnauthorizedError(err error) bool {
	var eventErr *EventError
	return errors.As(err, &eventErr) && eventErr.Code == "UNAUTHORIZED_ACCESS"
}

// GetEventErrorCode extracts the error code from an EventError
func GetEventErrorCode(err error) string {
	var eventErr *EventError
	if errors.As(err, &eventErr) {
		return eventErr.Code
	}
	return ""
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var eventErr *EventError
	if !errors.As(err, &eventErr) {
		return false
	}

	validationCodes := []string{
		"EVENT_DATE_INVALID",
		"TICKETS_EXCEED_CAPACITY",
		"INVALID_TICKET_REDUCTION",
		"CANNOT_UPDATE_CANCELLED",
		"CANNOT_UPDATE_COMPLETED",
		"CANNOT_CANCEL_COMPLETED",
		"CANNOT_DELETE_WITH_TICKETS",
		"EVENT_ALREADY_CANCELLED",
		"EVENT_ALREADY_COMPLETED",
	}

	for _, code := range validationCodes {
		if eventErr.Code == code {
			return true
		}
	}
	return false
}
