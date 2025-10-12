package session

import "errors"

var (
	ErrInvaliTokenHash   = errors.New("invalid token hash")
	ErrNotFound          = errors.New("session not found")
	ErrInvalidSession    = errors.New("invalid session")
	ErrInvalidExpiry     = errors.New("expiresAt must be in the future")
	ErrInvalidDeviceInfo = errors.New("invalid device info")
	ErrInvalidIP         = errors.New("invalid IP")
)
