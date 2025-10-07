package app

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

type PublicLinkService struct {
	publicLinkRepo domain.PublicLinkRepository
	queue          domain.PublicExpirer
}

func NewPublicLinkService() {

}

func (s *PublicLinkService) Create(fileID, createdByUserID uuid.UUID, tokenHash string, expiresAt time.Time) (*domain.PublicLink, error) {
	link := domain.NewPublicLink(fileID, createdByUserID, tokenHash, expiresAt)
	if err := s.queue.Enqueue(link.ID); err != nil {
		return nil, err
	}

	return link, nil
}

func Delete() {

}
