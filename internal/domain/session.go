package domain

import (
	"net"
	"time"

	uuid "github.com/google/uuid"
)

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID

	TokenHash        string
	RefreshTokenHash string
	DeviceInfo       string

	Ip net.IP

	IsRevoked bool

	LastUsedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  time.Time
}

func NewSession(
	userID uuid.UUID,
	tokenHash string,
	refreshTokenHash string,
	deviceInfo string,
	ip net.IP,
	expiresAt time.Time,
) *Session {
	now := time.Now()
	return &Session{
		ID:               uuid.New(),
		UserID:           userID,
		TokenHash:        tokenHash,
		RefreshTokenHash: refreshTokenHash,
		DeviceInfo:       deviceInfo,
		Ip:               ip,
		IsRevoked:        false,
		LastUsedAt:       now,
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        expiresAt,
	}
}

func (s *Session) Revoke() {
	s.IsRevoked = true
	s.UpdatedAt = time.Now()
}

func (s *Session) UpdateLastUsed() {
	s.LastUsedAt = time.Now()
	s.UpdatedAt = time.Now()
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
