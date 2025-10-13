package magic_link

import (
	"time"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type MagicLink struct {
	ID     uuid.UUID
	UserID uuid.UUID

	TokenHash  value_objects.TokenHash
	DeviceInfo value_objects.DeviceInfo
	Purpose    Purpose

	Ip value_objects.IP

	IsUsed    bool
	IsExpired bool

	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiredAt value_objects.ExpiresAt
}

func NewMagicLink(
	userID uuid.UUID,
	tokenHash value_objects.TokenHash,
	deviceInfo value_objects.DeviceInfo,
	purpose Purpose,
	ip value_objects.IP,
	expiresAt value_objects.ExpiresAt,
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
		ExpiredAt:  expiresAt,
	}
}

// MarkAsUsed помечает магическую ссылку как использованную
func (m *MagicLink) MarkAsUsed() {
	now := time.Now()
	m.IsUsed = true
	m.UsedAt = &now
	m.UpdatedAt = now
}

func (m *MagicLink) MarkAsExpired() error {
	now := time.Now()
	expiredAt, err := value_objects.NewExpiresAt(now)
	if err != nil {
		return err
	}

	m.IsExpired = true
	m.ExpiredAt = expiredAt
	m.UpdatedAt = now
	return nil
}
