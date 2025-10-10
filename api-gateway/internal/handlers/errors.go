package handlers

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func NewErrorResponse(errorCode, message string) ErrorResponse {
	return ErrorResponse{
		Error:   errorCode,
		Message: message,
	}
}

func NewErrorResponseWithDetails(errorCode, message string, details interface{}) ErrorResponse {
	return ErrorResponse{
		Error:   errorCode,
		Message: message,
		Details: details,
	}
}
