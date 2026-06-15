package service

import "fmt"

type Error struct {
	Status  int
	Code    string
	Message string
	Details any
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) StatusCode() int       { return e.Status }
func (e *Error) ErrorCode() string     { return e.Code }
func (e *Error) PublicMessage() string { return e.Message }
func (e *Error) ErrorDetails() any     { return e.Details }

func Validation(message string, details any) *Error {
	return &Error{Status: 400, Code: "VALIDATION_ERROR", Message: message, Details: details}
}
func Unauthorized(message string) *Error {
	return &Error{Status: 401, Code: "UNAUTHORIZED", Message: message}
}
func Forbidden(message string) *Error {
	return &Error{Status: 403, Code: "FORBIDDEN", Message: message}
}
func NotFound(message string) *Error { return &Error{Status: 404, Code: "NOT_FOUND", Message: message} }
func Conflict(message string) *Error { return &Error{Status: 409, Code: "CONFLICT", Message: message} }
func Internal(message string) *Error {
	return &Error{Status: 500, Code: "INTERNAL_ERROR", Message: message}
}
func ProviderUnavailable(message string) *Error {
	return &Error{Status: 400, Code: "PROVIDER_UNAVAILABLE", Message: message}
}
