package app

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/public_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

const defaultPublicLinkTTL = 15 * time.Minute

type PublicLinkService struct {
	queryRepo    public_link.PublicLinkQueryRepository
	commandRepo  public_link.PublicLinkCommandRepository
	queue        queue.Expirer
	eventService *EventService
}

func NewPublicLinkService(
	queryRepo public_link.PublicLinkQueryRepository,
	commandRepo public_link.PublicLinkCommandRepository,
	queue queue.Expirer,
	eventService *EventService,
) *PublicLinkService {
	return &PublicLinkService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		queue:        queue,
		eventService: eventService,
	}
}

func (s *PublicLinkService) Create(
	fileID, createdByUserID uuid.UUID,
	tokenHashRaw string,
	expiresAtRaw time.Time,
) (*public_link.PublicLink, error) {
	now := time.Now()

	tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
	if err != nil {
		return nil, err
	}

	expiresAt := expiresAtRaw
	if expiresAt.IsZero() {
		expiresAt = now.Add(defaultPublicLinkTTL)
	}
	expiresAtVO, err := value_objects.NewExpiresAt(expiresAt)
	if err != nil {
		return nil, err
	}

	link := public_link.NewPublicLink(fileID, createdByUserID, tokenHash.String(), expiresAtVO.Time())

	if err := s.commandRepo.Save(link); err != nil {
		return nil, err
	}

	ttl := time.Until(expiresAtVO.Time())
	if err := s.queue.Enqueue(link.ID, ttl); err != nil {
		return nil, err
	}

	if s.eventService != nil {
		eventName, payload := public_link.NewPublicLinkCreatedEvent(link)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return link, nil
}

func (s *PublicLinkService) Delete(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return public_link.ErrNotFound
	}

	if err := s.commandRepo.Delete(id); err != nil {
		return err
	}

	_ = s.queue.Remove(id)

	if s.eventService != nil {
		eventName, payload := public_link.NewPublicLinkDeletedEvent(id)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *PublicLinkService) Expire(id uuid.UUID) error {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if link == nil {
		return public_link.ErrNotFound
	}
	if link.IsExpired {
		return nil
	}

	link.MarkAsExpired()
	if err := s.commandRepo.Save(link); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := public_link.NewPublicLinkExpiredEvent(id)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *PublicLinkService) GetByID(id uuid.UUID) (*public_link.PublicLink, error) {
	link, err := s.queryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, public_link.ErrNotFound
	}
	if link.IsExpired {
		return nil, errors.New("link has expired")
	}
	return link, nil
}

func (s *PublicLinkService) GetAll() ([]*public_link.PublicLink, error) {
	return s.queryRepo.GetAll()
}
