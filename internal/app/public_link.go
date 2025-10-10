package app

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

const defaultPublicLinkTTL = 15 * time.Minute

type PublicLinkService struct {
	queryRepo   domain.PublicLinkQueryRepository
	commandRepo domain.PublicLinkCommandRepository
	queue       domain.Expirer
}

func NewPublicLinkService(
	queryRepo domain.PublicLinkQueryRepository,
	commandRepo domain.PublicLinkCommandRepository,
	queue domain.Expirer,
) *PublicLinkService {
	return &PublicLinkService{
		queryRepo:   queryRepo,
		commandRepo: commandRepo,
		queue:       queue,
	}
}

// --- Commands ---

func (s *PublicLinkService) Create(
	fileID, createdByUserID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) (*domain.PublicLink, error) {
	now := time.Now()
	if expiresAt.IsZero() {
		expiresAt = now.Add(defaultPublicLinkTTL)
	}
	if !expiresAt.After(now) {
		return nil, ErrInvalidExpiryTime
	}

	link := domain.NewPublicLink(fileID, createdByUserID, tokenHash, expiresAt)
	if err := s.commandRepo.Save(link); err != nil {
		return nil, err
	}

	ttl := time.Until(expiresAt)
	if err := s.queue.Enqueue(link.ID, ttl); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *PublicLinkService) Delete(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrPublicLinkNotFound
	}

	if err := s.commandRepo.Delete(id); err != nil {
		return err
	}

	_ = s.queue.Remove(id)
	return nil
}

func (s *PublicLinkService) Expire(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrPublicLinkNotFound
	}
	if link.IsExpired {
		return nil
	}

	link.MarkAsExpired()
	return s.commandRepo.Save(link)
}

// --- Queries ---

func (s *PublicLinkService) GetByID(id uuid.UUID) (*domain.PublicLink, error) {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrPublicLinkNotFound
	}
	if link.IsExpired {
		return nil, errors.New("link has expired")
	}
	return link, nil
}

func (s *PublicLinkService) GetAll() ([]*domain.PublicLink, error) {
	return s.queryRepo.GetAll()
}
