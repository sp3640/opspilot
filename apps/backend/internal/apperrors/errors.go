package apperrors

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")

	// User errors
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")

	// Generic errors
	ErrInternal = errors.New("internal server error")
)