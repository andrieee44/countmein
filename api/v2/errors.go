package api

import "errors"

var (
	ErrAuthRequired          = errors.New("authentication required")
	ErrInvalidPassword       = errors.New("password verification failed")
	ErrDisplayNameNotAllowed = errors.New("display names are not permitted for email signup")
	ErrSessionExpired        = errors.New("session or token has expired")
	ErrForbidden             = errors.New("action not permitted")
	ErrNotFoundOrDenied      = errors.New("resource not found or access denied")
)
