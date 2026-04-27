package api

import "errors"

var (
	ErrAuthRequired          error = errors.New("authentication required")
	ErrInvalidPassword       error = errors.New("password verification failed")
	ErrDisplayNameNotAllowed error = errors.New("display names are not permitted for email signup")
	ErrSessionExpired        error = errors.New("session or token has expired")
	ErrForbidden             error = errors.New("action not permitted")
	ErrNotFoundOrDenied      error = errors.New("resource not found or access denied")
)
