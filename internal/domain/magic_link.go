package domain

import (
	"net"
	"time"

	uuid "github.com/google/uuid"
)

type MagicLink struct {
	ID     uuid.UUID
	UserID uuid.UUID

	TokenHash  string
	DeviceInfo string
	Purpose    string

	Ip net.IP

	IsUsed    bool
	IsExpired bool

	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiredAt time.Time
}

func NewMagicLink(
	userID uuid.UUID,
	tokenHash string,
	deviceInfo string,
	purpose string,
	ip net.IP,
	expiresAt time.Time,
) *MagicLink {
	now := time.Now()
	return &MagicLink{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  tokenHash,
		DeviceInfo: deviceInfo,
		Purpose:    purpose,
		Ip:         ip,
		IsUsed:     false,
		UsedAt:     nil,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (m *MagicLink) MarkAsUsed() {
	now := time.Now()
	m.IsUsed = true
	m.UsedAt = &now
	m.UpdatedAt = now
}

func (m *MagicLink) MarkAsExpired() {
	now := time.Now()
	m.IsExpired = true
	m.ExpiredAt = now
	m.UpdatedAt = now
}
