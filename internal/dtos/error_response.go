package dtos

// ErrorResponse defines the standard error response according to the API contract.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
