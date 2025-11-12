package magic_link_service

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/app"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

const magicLinkTTL = 5 * time.Minute

type MagicLinkService struct {
	queryRepo    magic_link.QueryRepository
	commandRepo  magic_link.CommandRepository
	eventService *event_service.EventService
	uow          app.UnitOfWork // Добавлено
}

func NewMagicLinkService(
	queryRepo magic_link.QueryRepository,
	commandRepo magic_link.CommandRepository,
	eventService *event_service.EventService,
	uow app.UnitOfWork,
) *MagicLinkService {
	return &MagicLinkService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		eventService: eventService,
		uow:          uow,
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
	var createdLink *magic_link.MagicLink

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
		if err != nil {
			return err
		}

		deviceInfo, err := value_objects.NewDeviceInfo(deviceInfoRaw)
		if err != nil {
			return err
		}

		purpose, err := magic_link.NewPurpose(purposeRaw)
		if err != nil {
			return err
		}

		ipVO, err := value_objects.NewIP(ip)
		if err != nil {
			return err
		}

		expiresAtVO, err := value_objects.NewExpiresAt(time.Now().Add(magicLinkTTL))
		if err != nil {
			return err
		}

		link := magic_link.NewMagicLink(userID, tokenHash, deviceInfo, purpose, ipVO, expiresAtVO)

		if err := s.commandRepo.Save(ctx, link); err != nil {
			return err
		}

		createdLink = link
		return nil
	})

	if err != nil {
		return nil, err
	}

	if s.eventService != nil {
		eventName, payload := magic_link.NewMagicLinkCreatedEvent(createdLink)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return createdLink, nil
}

func (s *MagicLinkService) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	var link *magic_link.MagicLink

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		link, err = s.queryRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if link == nil {
			return magic_link.ErrNotFound
		}

		link.MarkAsUsed()

		return s.commandRepo.Save(ctx, link)
	})

	if err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := magic_link.NewMagicLinkUsedEvent(link)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

func (s *MagicLinkService) Delete(ctx context.Context, id uuid.UUID) error {
	var link *magic_link.MagicLink

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		link, err = s.queryRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if link == nil {
			return magic_link.ErrNotFound
		}

		return s.commandRepo.Delete(ctx, id)
	})

	if err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := magic_link.NewMagicLinkDeletedEvent(link)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

func (s *MagicLinkService) GetByID(ctx context.Context, id uuid.UUID) (*magic_link.MagicLink, error) {
	link, err := s.queryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, magic_link.ErrNotFound
	}
	return link, nil
}

func (s *MagicLinkService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*magic_link.MagicLink, error) {
	return s.queryRepo.GetByUserID(ctx, userID)
}

func (s *MagicLinkService) GetByTokenHash(ctx context.Context, tokenHashRaw string) (*magic_link.MagicLink, error) {
	tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
	if err != nil {
		return nil, err
	}

	link, err := s.queryRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, magic_link.ErrNotFound
	}

	if !link.IsValid() {
		return nil, magic_link.ErrInvalid
	}

	return link, nil
}

func (s *MagicLinkService) GetAll(ctx context.Context) ([]*magic_link.MagicLink, error) {
	return s.queryRepo.GetAll(ctx)
}

func (s *MagicLinkService) CleanupExpired(ctx context.Context) error {
	links, err := s.queryRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, link := range links {
		if link.IsExpired() {
			if err := s.Delete(ctx, link.ID); err != nil {
				continue
			}
		}
	}

	return nil
}
