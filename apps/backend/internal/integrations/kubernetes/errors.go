package kubernetes

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type ErrInvalidKubeconfig struct {
	Err error
}

func (e *ErrInvalidKubeconfig) Error() string {
	if e == nil || e.Err == nil {
		return "invalid kubeconfig"
	}

	return fmt.Sprintf("invalid kubeconfig: %v", e.Err)
}

func (e *ErrInvalidKubeconfig) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

type ErrConnectionFailed struct {
	Err error
}

func (e *ErrConnectionFailed) Error() string {
	if e == nil || e.Err == nil {
		return "connection failed"
	}

	return fmt.Sprintf("connection failed: %v", e.Err)
}

func (e *ErrConnectionFailed) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

type ErrAuthenticationFailed struct {
	Err error
}

func (e *ErrAuthenticationFailed) Error() string {
	if e == nil || e.Err == nil {
		return "authentication failed"
	}

	return fmt.Sprintf("authentication failed: %v", e.Err)
}

func (e *ErrAuthenticationFailed) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

type ErrAuthorizationFailed struct {
	Err error
}

func (e *ErrAuthorizationFailed) Error() string {
	if e == nil || e.Err == nil {
		return "authorization failed"
	}

	return fmt.Sprintf("authorization failed: %v", e.Err)
}

func (e *ErrAuthorizationFailed) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

type ErrVersionUnavailable struct {
	Err error
}

func (e *ErrVersionUnavailable) Error() string {
	if e == nil || e.Err == nil {
		return "version unavailable"
	}

	return fmt.Sprintf("version unavailable: %v", e.Err)
}

func (e *ErrVersionUnavailable) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func wrapConnectionError(err error) error {
	if err == nil {
		return nil
	}

	if apierrors.IsUnauthorized(err) {
		return &ErrAuthenticationFailed{Err: err}
	}
	if apierrors.IsForbidden(err) {
		return &ErrAuthorizationFailed{Err: err}
	}

	return &ErrConnectionFailed{Err: err}
}

func wrapVersionError(err error) error {
	if err == nil {
		return nil
	}

	if apierrors.IsUnauthorized(err) {
		return &ErrAuthenticationFailed{Err: err}
	}
	if apierrors.IsForbidden(err) {
		return &ErrAuthorizationFailed{Err: err}
	}

	return &ErrVersionUnavailable{Err: err}
}
