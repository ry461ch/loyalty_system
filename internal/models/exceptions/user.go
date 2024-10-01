package exceptions

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAuthentication = errors.New("user unauthorized")
	ErrUserConflict       = errors.New("user already exists")
	ErrUserBadFormat      = errors.New("user bad data format")
)
