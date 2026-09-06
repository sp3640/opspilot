package github

import "errors"

// Typed errors this package's Client can return - mapped to the
// corresponding apperrors.ErrGitHub* at the service boundary. Kept
// package-local (rather than importing apperrors here) so this package has
// no dependency on the rest of the backend and stays independently
// testable/reusable.
var (
	ErrUnauthorized  = errors.New("github: invalid or revoked credentials")
	ErrForbidden     = errors.New("github: access denied")
	ErrRateLimited   = errors.New("github: rate limit exceeded")
	ErrNotFound      = errors.New("github: resource not found")
	ErrUnavailable   = errors.New("github: service unavailable")
	ErrTimeout       = errors.New("github: request timed out")
	ErrTokenRequired = errors.New("github: an access token is required")
)
