package errors

type APIError struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func New(code, message string, details interface{}) APIError {
	return APIError{Error: code, Message: message, Details: details}
}
