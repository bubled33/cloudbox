package domain

import (
	"time"

	"github.com/google/uuid"
)

type PublicLink struct {
	ID              uuid.UUID
	FileID          uuid.UUID
	CreatedByUserID uuid.UUID

	TokenHash string
	IsExpired bool

	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiredAt time.Time
}

func NewPublicLink(
	fileID uuid.UUID,
	createdByUserID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) *PublicLink {
	now := time.Now()

	return &PublicLink{
		ID:              uuid.New(),
		FileID:          fileID,
		CreatedByUserID: createdByUserID,
		TokenHash:       tokenHash,
		IsExpired:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (p *PublicLink) MarkAsExpired() {
	now := time.Now()
	p.IsExpired = true
	p.ExpiredAt = now
	p.UpdatedAt = now
}
