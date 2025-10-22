package app

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

const magicLinkTTL = 5 * time.Minute

type MagicLinkService struct {
	queryRepo       magic_link.QueryRepository
	commandRepo     magic_link.CommandRepository
	expirerProducer queue.ExpirerProducer
	expirerConsumer queue.ExpirerConsumer
	eventService    *EventService
}

func NewMagicLinkService(
	queryRepo magic_link.QueryRepository,
	commandRepo magic_link.CommandRepository,
	expirerProducer queue.ExpirerProducer,
	expirerConsumer queue.ExpirerConsumer,
	eventService *EventService,
) *MagicLinkService {
	return &MagicLinkService{
		queryRepo:       queryRepo,
		commandRepo:     commandRepo,
		expirerProducer: expirerProducer,
		eventService:    eventService,
	}
}

func (s *MagicLinkService) Create(
	ctx context.Context,
	userID uuid.UUID,
	tokenHashRaw string,
	deviceInfoRaw string,
	purposeRaw string,
	ip net.IP,
) (*magic_link.MagicLink, error) {
	tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
	if err != nil {
		return nil, err
	}

	deviceInfo, err := value_objects.NewDeviceInfo(deviceInfoRaw)
	if err != nil {
		return nil, err
	}

	purpose, err := magic_link.NewPurpose(purposeRaw)
	if err != nil {
		return nil, err
	}

	ipVO, err := value_objects.NewIP(ip)
	if err != nil {
		return nil, err
	}

	expiresAtVO, err := value_objects.NewExpiresAt(time.Now().Add(magicLinkTTL))
	if err != nil {
		return nil, err
	}

	link := magic_link.NewMagicLink(userID, tokenHash, deviceInfo, purpose, ipVO, expiresAtVO)

	if err := s.commandRepo.Save(link); err != nil {
		return nil, err
	}

	if err := s.expirerProducer.Produce(ctx, link.ID, magicLinkTTL); err != nil {
		return nil, err
	}

	if s.eventService != nil {
		eventName, payload := magic_link.NewMagicLinkCreatedEvent(link)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return link, nil
}

func (s *MagicLinkService) MarkAsUsed(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return magic_link.ErrNotFound
	}

	link.MarkAsUsed()

	if err := s.commandRepo.Save(link); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := magic_link.NewMagicLinkUsedEvent(link)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *MagicLinkService) Expire(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return magic_link.ErrNotFound
	}

	if err := link.MarkAsExpired(); err != nil {
		return err
	}

	if err := s.commandRepo.Save(link); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := magic_link.NewMagicLinkExpiredEvent(link)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *MagicLinkService) Delete(ctx context.Context, id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return magic_link.ErrNotFound
	}

	if err := s.commandRepo.Delete(id); err != nil {
		return err
	}

	_ = s.expirerConsumer.Remove(ctx, id)

	if s.eventService != nil {
		eventName, payload := magic_link.NewMagicLinkDeletedEvent(link)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *MagicLinkService) GetByID(id uuid.UUID) (*magic_link.MagicLink, error) {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, magic_link.ErrNotFound
	}
	return link, nil
}

func (s *MagicLinkService) GetByUserID(userID uuid.UUID) ([]*magic_link.MagicLink, error) {
	return s.queryRepo.GetByUserID(userID)
}

func (s *MagicLinkService) GetByTokenHash(tokenHashRaw string) (*magic_link.MagicLink, error) {
	tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
	if err != nil {
		return nil, err
	}

	link, err := s.queryRepo.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, magic_link.ErrNotFound
	}
	return link, nil
}

func (s *MagicLinkService) GetAll() ([]*magic_link.MagicLink, error) {
	return s.queryRepo.GetAll()
}

func (s *MagicLinkService) CleanupExpired(ctx context.Context) error {
	links, err := s.queryRepo.GetAll()
	if err != nil {
		return err
	}

	for _, link := range links {
		if link.IsExpired {
			if err := s.Delete(ctx, link.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
