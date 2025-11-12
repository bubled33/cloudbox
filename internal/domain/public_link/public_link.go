package public_link

import (
	"time"

	"github.com/google/uuid"
)

type PublicLink struct {
	ID              uuid.UUID
	FileID          uuid.UUID
	CreatedByUserID uuid.UUID

	TokenHash string

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
		ExpiredAt:       expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (p *PublicLink) IsExpired() bool {
	return time.Now().After(p.ExpiredAt)
}
