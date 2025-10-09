package app

import (
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

var ErrMagicLinkNotFound = errors.New("magic link not found")

const magicLinkTTL = 5 * time.Minute

type MagicLinkService struct {
	magicLinkRepo domain.MagicLinkRepository
	queue         domain.Expirer
}

func NewMagicLinkService(
	magicLinkRepo domain.MagicLinkRepository,
	queue domain.Expirer,
) *MagicLinkService {
	return &MagicLinkService{
		magicLinkRepo: magicLinkRepo,
		queue:         queue,
	}
}

// Создание новой магической ссылки
func (s *MagicLinkService) Create(
	userID uuid.UUID,
	tokenHash string,
	deviceInfo string,
	purpose string,
	ip net.IP,
) (*domain.MagicLink, error) {
	expiresAt := time.Now().Add(magicLinkTTL)
	link := domain.NewMagicLink(userID, tokenHash, deviceInfo, purpose, ip, expiresAt)

	if err := s.magicLinkRepo.Save(link); err != nil {
		return nil, err
	}

	if err := s.queue.Enqueue(link.ID, magicLinkTTL); err != nil {
		return nil, err
	}

	return link, nil
}

// Получение ссылки по ID
func (s *MagicLinkService) GetByID(id uuid.UUID) (*domain.MagicLink, error) {
	link, err := s.magicLinkRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrMagicLinkNotFound
	}
	return link, nil
}

// Получение ссылок по пользователю
func (s *MagicLinkService) GetByUserID(userID uuid.UUID) ([]*domain.MagicLink, error) {
	return s.magicLinkRepo.GetByUserID(userID)
}

// Получение ссылки по токену
func (s *MagicLinkService) GetByTokenHash(tokenHash string) (*domain.MagicLink, error) {
	link, err := s.magicLinkRepo.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrMagicLinkNotFound
	}
	return link, nil
}

// Получение всех ссылок
func (s *MagicLinkService) GetAll() ([]*domain.MagicLink, error) {
	return s.magicLinkRepo.GetAll()
}

// Маркировка ссылки как использованной
func (s *MagicLinkService) MarkAsUsed(id uuid.UUID) error {
	link, err := s.magicLinkRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrMagicLinkNotFound
	}

	link.MarkAsUsed()
	return s.magicLinkRepo.Save(link)
}

// Истечение ссылки (автоистечение)
func (s *MagicLinkService) Expire(id uuid.UUID) error {
	link, err := s.magicLinkRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrMagicLinkNotFound
	}

	link.MarkAsExpired()

	return s.magicLinkRepo.Save(link)
}

// Удаление ссылки
func (s *MagicLinkService) Delete(id uuid.UUID) error {
	link, err := s.magicLinkRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrMagicLinkNotFound
	}

	if err := s.magicLinkRepo.Delete(id); err != nil {
		return err
	}

	_ = s.queue.Remove(id)
	return nil
}

// Очистка всех просроченных ссылок
func (s *MagicLinkService) CleanupExpired() error {
	links, err := s.magicLinkRepo.GetAll()
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
