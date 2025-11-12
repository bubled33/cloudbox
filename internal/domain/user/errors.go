package user

import "errors"

var (
	ErrNotFound               = errors.New("user not found")
	ErrInvalidEmailFormat     = errors.New("invalid email format")
	ErrInvalidDisplayNameSize = errors.New("display name must be between 2 and 50 characters")
)
