package app

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

var (
	ErrPublicLinkNotFound = errors.New("public link not found")
	ErrInvalidExpiryTime  = errors.New("expiresAt must be in the future")
)

const defaultPublicLinkTTL = 15 * time.Minute

type PublicLinkService struct {
	publicLinkRepo domain.PublicLinkRepository
	queue          domain.Expirer
}

func NewPublicLinkService(repo domain.PublicLinkRepository, queue domain.Expirer) *PublicLinkService {
	return &PublicLinkService{
		publicLinkRepo: repo,
		queue:          queue,
	}
}

func (s *PublicLinkService) Create(
	fileID, createdByUserID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) (*domain.PublicLink, error) {
	now := time.Now()

	// если не передано — задаём дефолт
	if expiresAt.IsZero() {
		expiresAt = now.Add(defaultPublicLinkTTL)
	}

	if !expiresAt.After(now) {
		return nil, ErrInvalidExpiryTime
	}

	link := domain.NewPublicLink(fileID, createdByUserID, tokenHash, expiresAt)

	if err := s.publicLinkRepo.Save(link); err != nil {
		return nil, err
	}

	ttl := time.Until(expiresAt)
	if err := s.queue.Enqueue(link.ID, ttl); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *PublicLinkService) GetByID(id uuid.UUID) (*domain.PublicLink, error) {
	link, err := s.publicLinkRepo.GetByID(id)
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
	return s.publicLinkRepo.GetAll()
}

func (s *PublicLinkService) Delete(id uuid.UUID) error {
	link, err := s.publicLinkRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrPublicLinkNotFound
	}

	if err := s.publicLinkRepo.Delete(id); err != nil {
		return err
	}

	_ = s.queue.Remove(id)
	return nil
}

func (s *PublicLinkService) Expire(id uuid.UUID) error {
	link, err := s.publicLinkRepo.GetByID(id)
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
	return s.publicLinkRepo.Save(link)
}
