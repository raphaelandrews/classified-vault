package apperr

import "fmt"

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("apperror: code=%d msg=%s", e.Code, e.Message)
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func BadRequest(msg string) *AppError   { return New(400, msg) }
func Unauthorized(msg string) *AppError { return New(401, msg) }
func Forbidden(msg string) *AppError    { return New(403, msg) }
func NotFound(msg string) *AppError     { return New(404, msg) }
func Internal(msg string) *AppError     { return New(500, msg) }
