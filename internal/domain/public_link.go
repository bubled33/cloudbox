package domain

import (
	"time"

	uuid "github.com/google/uuid"
)

type PublicLink struct {
	ID              uuid.UUID
	FileID          uuid.UUID
	CreatedByUserID uuid.UUID

	TokenHash string
	IsExpired bool

	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
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
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       expiresAt,
	}
}

func (p *PublicLink) MarkAsExpired() {
	p.ExpiresAt = time.Now()
	p.IsExpired = true
}
