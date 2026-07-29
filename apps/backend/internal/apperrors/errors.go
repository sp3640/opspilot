package apperrors

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")

	// User errors
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")

	// Project errors
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectForbidden     = errors.New("forbidden")
	ErrProjectAlreadyExists = errors.New("project already exists")

	// Incident errors
	ErrIncidentNotFound = errors.New("incident not found")
	ErrInvalidSeverity  = errors.New("invalid severity")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrInvalidProject   = errors.New("invalid project")

	// Comment errors
	ErrCommentNotFound       = errors.New("comment not found")
	ErrCommentForbidden      = errors.New("forbidden")
	ErrInvalidCommentContent = errors.New("invalid comment content")

	// Generic errors
	ErrInternal = errors.New("internal server error")
)
