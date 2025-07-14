package common

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents the standard success response structure
type SuccessResponse struct {
	Message string `json:"message"`
}

// ListResponse represents a generic list response structure
type ListResponse[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}
