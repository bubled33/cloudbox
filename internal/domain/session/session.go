package session

import (
	"time"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID

	TokenHash        value_objects.TokenHash
	RefreshTokenHash value_objects.TokenHash
	DeviceInfo       value_objects.DeviceInfo
	Ip               value_objects.IP

	IsRevoked bool

	LastUsedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  value_objects.ExpiresAt
}

func NewSession(
	userID uuid.UUID,
	tokenHash value_objects.TokenHash,
	refreshTokenHash value_objects.TokenHash,
	deviceInfo value_objects.DeviceInfo,
	ip value_objects.IP,
	expiresAt value_objects.ExpiresAt,
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
	return s.ExpiresAt.IsExpired()
}
