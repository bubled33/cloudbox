package session_service

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/app"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type SessionService struct {
	queryRepo    session.QueryRepository
	commandRepo  session.CommandRepository
	eventService *event_service.EventService
	uow          app.UnitOfWork // Добавлено
}

func NewSessionService(
	queryRepo session.QueryRepository,
	commandRepo session.CommandRepository,
	eventService *event_service.EventService,
	uow app.UnitOfWork,
) *SessionService {
	return &SessionService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		eventService: eventService,
		uow:          uow,
	}
}

func (s *SessionService) Create(
	ctx context.Context,
	userID uuid.UUID,
	tokenHashRaw string,
	refreshTokenHashRaw string,
	deviceInfoRaw string,
	ip net.IP,
	expiresAt time.Time,
) (*session.Session, error) {
	var createdSession *session.Session

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		tokenHash, err := value_objects.NewTokenHash(tokenHashRaw)
		if err != nil {
			return err
		}

		refreshTokenHash, err := value_objects.NewTokenHash(refreshTokenHashRaw)
		if err != nil {
			return err
		}

		deviceInfo, err := value_objects.NewDeviceInfo(deviceInfoRaw)
		if err != nil {
			return err
		}

		ipVO, err := value_objects.NewIP(ip)
		if err != nil {
			return err
		}

		expiresAtVO, err := value_objects.NewExpiresAt(expiresAt)
		if err != nil {
			return err
		}

		sess := session.NewSession(userID, tokenHash, refreshTokenHash, deviceInfo, ipVO, expiresAtVO)

		if err := s.commandRepo.Save(ctx, sess); err != nil {
			return err
		}

		createdSession = sess
		return nil
	})

	if err != nil {
		return nil, err
	}

	if s.eventService != nil {
		eventName, payload := session.NewSessionCreatedEvent(createdSession)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return createdSession, nil
}

func (s *SessionService) Delete(ctx context.Context, sessionID uuid.UUID) error {
	var sess *session.Session

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		sess, err = s.queryRepo.GetByID(ctx, sessionID)
		if err != nil {
			return err
		}
		if sess == nil {
			return session.ErrNotFound
		}

		return s.commandRepo.Delete(ctx, sessionID)
	})

	if err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := session.NewSessionDeletedEvent(sessionID)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

func (s *SessionService) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	var sess *session.Session

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		sess, err = s.queryRepo.GetByID(ctx, sessionID)
		if err != nil {
			return err
		}
		if sess == nil {
			return session.ErrNotFound
		}

		sess.Revoke()
		return s.commandRepo.Save(ctx, sess)
	})

	if err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := session.NewSessionRevokedEvent(sessionID)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

func (s *SessionService) CleanupExpired(ctx context.Context) error {
	sessions, err := s.queryRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		if sess.IsExpired() {
			err := s.uow.Do(ctx, func(ctx context.Context) error {
				sess.Revoke()
				return s.commandRepo.Save(ctx, sess)
			})

			if err != nil {
				continue
			}

			if s.eventService != nil {
				eventName, payload := session.NewSessionExpiredEvent(sess.ID)
				_, _ = s.eventService.Create(ctx, eventName, payload)
			}
		}
	}

	return nil
}

func (s *SessionService) GetByAccessToken(ctx context.Context, accessToken string) (*session.Session, error) {
	tokenHash, err := value_objects.NewTokenHash(accessToken)
	if err != nil {
		return nil, err
	}

	sess, err := s.queryRepo.GetByAccessToken(ctx, &tokenHash)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, session.ErrNotFound
	}
	return sess, nil
}

func (s *SessionService) Touch(ctx context.Context, sessionID uuid.UUID) error {
	return s.uow.Do(ctx, func(ctx context.Context) error {
		sess, err := s.queryRepo.GetByID(ctx, sessionID)
		if err != nil {
			return err
		}
		if sess == nil {
			return session.ErrNotFound
		}

		sess.UpdateLastUsed()
		return s.commandRepo.Save(ctx, sess)
	})
}

func (s *SessionService) GetByID(ctx context.Context, sessionID uuid.UUID) (*session.Session, error) {
	sess, err := s.queryRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, session.ErrNotFound
	}
	return sess, nil
}

func (s *SessionService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*session.Session, error) {
	return s.queryRepo.GetByUserID(ctx, userID)
}

func (s *SessionService) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.queryRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		if !sess.IsRevoked {
			if err := s.Revoke(ctx, sess.ID); err != nil {

				continue
			}
		}
	}

	return nil
}

func (s *SessionService) RefreshAccessToken(ctx context.Context, refreshTokenRaw string) (*session.Session, error) {
	refreshTokenHash, err := value_objects.NewTokenHash(refreshTokenRaw)
	if err != nil {
		return nil, err
	}

	sess, err := s.queryRepo.GetByRefreshToken(ctx, &refreshTokenHash)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, session.ErrNotFound
	}

	if sess.IsRevoked || sess.IsExpired() {
		return nil, session.ErrInvalidSession
	}

	return sess, nil
}
