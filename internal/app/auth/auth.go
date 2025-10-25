package auth_service

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	magic_link_service "github.com/yourusername/cloud-file-storage/internal/app/magic_link"
	session_service "github.com/yourusername/cloud-file-storage/internal/app/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
)

type AuthService struct {
	magicLinkService *magic_link_service.MagicLinkService
	sessionService   *session_service.SessionService
	sessionTTL       time.Duration
}

func NewAuthService(mlService *magic_link_service.MagicLinkService, sService *session_service.SessionService, sessionTTL time.Duration) *AuthService {
	return &AuthService{
		magicLinkService: mlService,
		sessionService:   sService,
		sessionTTL:       sessionTTL,
	}
}

func (a *AuthService) RequestMagicLink(ctx context.Context, userID uuid.UUID, deviceInfo string, ip net.IP) (*magic_link.MagicLink, error) {
	tokenHash := generateTokenHash()
	return a.magicLinkService.Create(ctx, userID, tokenHash, deviceInfo, "login", ip)
}

func (a *AuthService) Authenticate(tokenHash string, ip net.IP) (*session.Session, error) {
	link, err := a.magicLinkService.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, magic_link.ErrMagicLink
	}

	// Используем метод IsValid() вместо проверки отдельных флагов
	if !link.IsValid() {
		return nil, magic_link.ErrMagicLink
	}

	if err := a.magicLinkService.MarkAsUsed(link.ID); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(a.sessionTTL)
	session, err := a.sessionService.Create(
		link.UserID,
		generateTokenHash(),
		generateTokenHash(),
		link.DeviceInfo.String(),
		ip,
		expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (a *AuthService) ValidateSession(sessionID uuid.UUID) (*session.Session, error) {
	sess, err := a.sessionService.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	if sess.IsRevoked || sess.IsExpired() {
		return nil, session.ErrInvalidSession
	}

	if err := a.sessionService.Touch(sessionID); err != nil {
		return nil, err
	}

	return sess, nil
}

func (a *AuthService) RevokeSession(sessionID uuid.UUID) error {
	return a.sessionService.Revoke(sessionID)
}

func generateTokenHash() string {
	return uuid.New().String()
}
