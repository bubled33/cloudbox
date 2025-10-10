package app

import (
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

const magicLinkTTL = 5 * time.Minute

type MagicLinkService struct {
	queryRepo   domain.MagicLinkQueryRepository
	commandRepo domain.MagicLinkCommandRepository
	queue       domain.Expirer
}

func NewMagicLinkService(
	queryRepo domain.MagicLinkQueryRepository,
	commandRepo domain.MagicLinkCommandRepository,
	queue domain.Expirer,
) *MagicLinkService {
	return &MagicLinkService{
		queryRepo:   queryRepo,
		commandRepo: commandRepo,
		queue:       queue,
	}
}

// --- Commands ---

func (s *MagicLinkService) Create(
	userID uuid.UUID,
	tokenHash string,
	deviceInfo string,
	purpose string,
	ip net.IP,
) (*domain.MagicLink, error) {
	expiresAt := time.Now().Add(magicLinkTTL)
	link := domain.NewMagicLink(userID, tokenHash, deviceInfo, purpose, ip, expiresAt)

	if err := s.commandRepo.Save(link); err != nil {
		return nil, err
	}

	if err := s.queue.Enqueue(link.ID, magicLinkTTL); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *MagicLinkService) MarkAsUsed(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrMagicLinkNotFound
	}

	link.MarkAsUsed()
	return s.commandRepo.Save(link)
}

func (s *MagicLinkService) Expire(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrMagicLinkNotFound
	}

	link.MarkAsExpired()
	return s.commandRepo.Save(link)
}

func (s *MagicLinkService) Delete(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrMagicLinkNotFound
	}

	if err := s.commandRepo.Delete(id); err != nil {
		return err
	}

	_ = s.queue.Remove(id)
	return nil
}

// --- Queries ---

func (s *MagicLinkService) GetByID(id uuid.UUID) (*domain.MagicLink, error) {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrMagicLinkNotFound
	}
	return link, nil
}

func (s *MagicLinkService) GetByUserID(userID uuid.UUID) ([]*domain.MagicLink, error) {
	return s.queryRepo.GetByUserID(userID)
}

func (s *MagicLinkService) GetByTokenHash(tokenHash string) (*domain.MagicLink, error) {
	link, err := s.queryRepo.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrMagicLinkNotFound
	}
	return link, nil
}

func (s *MagicLinkService) GetAll() ([]*domain.MagicLink, error) {
	return s.queryRepo.GetAll()
}

func (s *MagicLinkService) CleanupExpired() error {
	links, err := s.queryRepo.GetAll()
	if err != nil {
		return err
	}

	for _, link := range links {
		if link.IsExpired {
			if err := s.Delete(link.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
