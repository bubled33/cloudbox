package public_link_service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/app"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/public_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

const defaultPublicLinkTTL = 15 * time.Minute

type PublicLinkService struct {
	queryRepo    public_link.PublicLinkQueryRepository
	commandRepo  public_link.PublicLinkCommandRepository
	eventService *event_service.EventService
	uow          app.UnitOfWork
}

func NewPublicLinkService(
	queryRepo public_link.PublicLinkQueryRepository,
	commandRepo public_link.PublicLinkCommandRepository,
	eventService *event_service.EventService,
	uow app.UnitOfWork,
) *PublicLinkService {
	return &PublicLinkService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		eventService: eventService,
		uow:          uow,
	}
}

// Create создаёт новую публичную ссылку в транзакции
func (s *PublicLinkService) Create(
	ctx context.Context,
	fileID, createdByUserID uuid.UUID,
	tokenHashRaw string,
	expiresAtRaw time.Time,
) (*public_link.PublicLink, error) {
	var link *public_link.PublicLink

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		now := time.Now()

		tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
		if err != nil {
			return err
		}

		expiresAt := expiresAtRaw
		if expiresAt.IsZero() {
			expiresAt = now.Add(defaultPublicLinkTTL)
		}
		expiresAtVO, err := value_objects.NewExpiresAt(expiresAt)
		if err != nil {
			return err
		}

		link = public_link.NewPublicLink(
			fileID,
			createdByUserID,
			tokenHash.String(),
			expiresAtVO.Time(),
		)

		return s.commandRepo.Save(ctx, link)
	})

	if err != nil {
		return nil, err
	}

	if s.eventService != nil {
		eventName, payload := public_link.NewPublicLinkCreatedEvent(link)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return link, nil
}

// Delete удаляет публичную ссылку в транзакции
func (s *PublicLinkService) Delete(ctx context.Context, id uuid.UUID) error {
	var link *public_link.PublicLink

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		link, err = s.queryRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if link == nil {
			return public_link.ErrNotFound
		}

		return s.commandRepo.Delete(ctx, id)
	})

	if err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := public_link.NewPublicLinkDeletedEvent(id)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

// GetByID получает ссылку по ID с проверкой срока действия
func (s *PublicLinkService) GetByID(ctx context.Context, id uuid.UUID) (*public_link.PublicLink, error) {
	link, err := s.queryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, public_link.ErrNotFound
	}
	if link.IsExpired() {
		return nil, errors.New("link has expired")
	}
	return link, nil
}

// GetByID получает ссылку по ID с проверкой срока действия
func (s *PublicLinkService) GetByToken(ctx context.Context, token string) (*public_link.PublicLink, error) {
	link, err := s.queryRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, public_link.ErrNotFound
	}
	if link.IsExpired() {
		return nil, errors.New("link has expired")
	}
	return link, nil
}

// GetAll получает все публичные ссылки
func (s *PublicLinkService) GetAll(ctx context.Context) ([]*public_link.PublicLink, error) {
	return s.queryRepo.GetAll(ctx)
}

// GetByFileID получает все активные ссылки для файла
func (s *PublicLinkService) GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*public_link.PublicLink, error) {
	links, err := s.queryRepo.GetByFileID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	active := make([]*public_link.PublicLink, 0, len(links))
	for _, l := range links {
		if l != nil && !l.IsExpired() {
			active = append(active, l)
		}
	}

	return active, nil
}
