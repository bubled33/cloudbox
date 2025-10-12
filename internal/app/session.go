package app

import (
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type SessionService struct {
	queryRepo    session.QueryRepository
	commandRepo  session.CommandRepository
	eventService *EventService
}

func NewSessionService(
	queryRepo session.QueryRepository,
	commandRepo session.CommandRepository,
	eventService *EventService,
) *SessionService {
	return &SessionService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		eventService: eventService,
	}
}

// --- Commands ---

func (s *SessionService) Create(
	userID uuid.UUID,
	tokenHashRaw string,
	refreshTokenHashRaw string,
	deviceInfoRaw string,
	ip net.IP,
	expiresAt time.Time,
) (*session.Session, error) {
	// создаём Value Objects
	tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
	if err != nil {
		return nil, err
	}

	refreshTokenHash, err := value_objects.NewTokenHash(refreshTokenHashRaw)
	if err != nil {
		return nil, err
	}

	deviceInfo, err := value_objects.NewDeviceInfo(deviceInfoRaw)
	if err != nil {
		return nil, err
	}

	ipVO, err := value_objects.NewIP(ip)
	if err != nil {
		return nil, err
	}

	expiresAtVO, err := value_objects.NewExpiresAt(expiresAt)
	if err != nil {
		return nil, err
	}

	// создаём сессию
	sess := session.NewSession(userID, tokenHash, refreshTokenHash, deviceInfo, ipVO, expiresAtVO)

	if err := s.commandRepo.Save(sess); err != nil {
		return nil, err
	}

	// Событие создания сессии
	if s.eventService != nil {
		eventName, payload := session.NewSessionCreatedEvent(sess)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return sess, nil
}

func (s *SessionService) Delete(sessionID uuid.UUID) error {
	sess, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if sess == nil {
		return session.ErrNotFound
	}

	if err := s.commandRepo.Delete(sessionID); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := session.NewSessionDeletedEvent(sessionID)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *SessionService) Revoke(sessionID uuid.UUID) error {
	sess, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if sess == nil {
		return session.ErrNotFound
	}

	sess.Revoke()
	if err := s.commandRepo.Save(sess); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := session.NewSessionRevokedEvent(sessionID)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *SessionService) CleanupExpired() error {
	sessions, err := s.queryRepo.GetAll()
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		if sess.IsExpired() {
			sess.Revoke()
			if err := s.commandRepo.Save(sess); err != nil {
				return err
			}

			if s.eventService != nil {
				eventName, payload := session.NewSessionExpiredEvent(sess.ID)
				_, _ = s.eventService.Create(eventName, payload)
			}
		}
	}

	return nil
}

func (s *SessionService) Touch(sessionID uuid.UUID) error {
	sess, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if sess == nil {
		return session.ErrNotFound
	}

	sess.UpdateLastUsed()
	return s.commandRepo.Save(sess)
}

// --- Queries ---

func (s *SessionService) GetByID(sessionID uuid.UUID) (*session.Session, error) {
	sess, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, session.ErrNotFound
	}
	return sess, nil
}

func (s *SessionService) GetByUserID(userID uuid.UUID) ([]*session.Session, error) {
	return s.queryRepo.GetByUserID(userID)
}
