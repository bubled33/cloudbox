package domainerrors

import "errors"

var (
	ErrInvaliTokenHash     = errors.New("invalid token hash")
	ErrInvalidExpiry       = errors.New("expiresAt must be in the future")
	ErrInvalidDeviceInfo   = errors.New("invalid device info")
	ErrInvalidIP           = errors.New("invalid IP")
	ErrTransactionNotFound = errors.New("transactio not found")
)
