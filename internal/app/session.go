package app

import (
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

type SessionService struct {
	queryRepo   domain.SessionQueryRepository
	commandRepo domain.SessionCommandRepository
}

func NewSessionService(
	queryRepo domain.SessionQueryRepository,
	commandRepo domain.SessionCommandRepository,
) *SessionService {
	return &SessionService{
		queryRepo:   queryRepo,
		commandRepo: commandRepo,
	}
}

// --- Commands ---

func (s *SessionService) Create(
	userID uuid.UUID,
	tokenHash string,
	refreshTokenHash string,
	deviceInfo string,
	ip net.IP,
	expiresAt time.Time,
) (*domain.Session, error) {
	if time.Until(expiresAt) <= 0 {
		return nil, ErrInvalidExpiry
	}

	session := domain.NewSession(userID, tokenHash, refreshTokenHash, deviceInfo, ip, expiresAt)

	if err := s.commandRepo.Save(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) Delete(sessionID uuid.UUID) error {
	session, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	return s.commandRepo.Delete(sessionID)
}

func (s *SessionService) Revoke(sessionID uuid.UUID) error {
	session, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	session.Revoke()
	return s.commandRepo.Save(session)
}

func (s *SessionService) CleanupExpired() error {
	sessions, err := s.queryRepo.GetAll()
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.IsExpired() {
			session.Revoke()
			if err := s.commandRepo.Save(session); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SessionService) Touch(sessionID uuid.UUID) error {
	session, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	session.UpdateLastUsed()
	return s.commandRepo.Save(session)
}

// --- Queries ---

func (s *SessionService) GetByID(sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.queryRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *SessionService) GetByUserID(userID uuid.UUID) ([]*domain.Session, error) {
	return s.queryRepo.GetByUserID(userID)
}
