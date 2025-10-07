package app

import (
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionService struct {
	sessionRepo domain.SessionRepository
}

func NewSessionService(sessionRepo domain.SessionRepository) *SessionService {
	return &SessionService{sessionRepo: sessionRepo}
}

func (s *SessionService) Create(
	userID uuid.UUID,
	tokenHash string,
	refreshTokenHash string,
	deviceInfo string,
	ip net.IP,
	expiresAt time.Time,
) (*domain.Session, error) {
	if time.Until(expiresAt) <= 0 {
		return nil, errors.New("expiresAt must be in the future")
	}

	session := domain.NewSession(userID, tokenHash, refreshTokenHash, deviceInfo, ip, expiresAt)
	if err := s.sessionRepo.Save(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) Delete(sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	return s.sessionRepo.Delete(sessionID)
}

func (s *SessionService) GetByID(sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *SessionService) GetByUserID(userID uuid.UUID) ([]*domain.Session, error) {
	return s.sessionRepo.GetByUserID(userID)
}

func (s *SessionService) Revoke(sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	return s.sessionRepo.Revoke(sessionID)
}

func (s *SessionService) CleanupExpired() error {
	sessions, err := s.sessionRepo.GetAll()
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if time.Now().After(session.ExpiresAt) {
			if err := s.sessionRepo.Delete(session.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
