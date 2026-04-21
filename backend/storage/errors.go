package storage

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrSessionExpired     = errors.New("session expired")
	ErrSiteNotFound       = errors.New("site not found")
	ErrSiteAlreadyExists  = errors.New("site already exists")
)
