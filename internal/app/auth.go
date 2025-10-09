package app

import (
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

var (
	ErrInvalidMagicLink = errors.New("invalid or expired magic link")
)

type AuthService struct {
	magicLinkService *MagicLinkService
	sessionService   *SessionService
	sessionTTL       time.Duration
}

func NewAuthService(mlService *MagicLinkService, sService *SessionService, sessionTTL time.Duration) *AuthService {
	return &AuthService{
		magicLinkService: mlService,
		sessionService:   sService,
		sessionTTL:       sessionTTL,
	}
}

func (a *AuthService) RequestMagicLink(userID uuid.UUID, deviceInfo string, ip net.IP) (*domain.MagicLink, error) {
	tokenHash := generateTokenHash()

	link, err := a.magicLinkService.Create(userID, tokenHash, deviceInfo, "login", ip)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (a *AuthService) Authenticate(tokenHash string, ip net.IP) (*domain.Session, error) {
	link, err := a.magicLinkService.GetByTokenHash(tokenHash)
	if err != nil || link.IsUsed || link.IsExpired {
		return nil, ErrInvalidMagicLink
	}

	if err := a.magicLinkService.MarkAsUsed(link.ID); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(a.sessionTTL)
	session, err := a.sessionService.Create(link.UserID, generateTokenHash(), generateTokenHash(), link.DeviceInfo, ip, expiresAt)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (a *AuthService) ValidateSession(sessionID uuid.UUID) (*domain.Session, error) {
	session, err := a.sessionService.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsRevoked || session.IsExpired() {
		return nil, errors.New("invalid session")
	}

	session.UpdateLastUsed()
	if err := a.sessionService.sessionRepo.Save(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (a *AuthService) RevokeSession(sessionID uuid.UUID) error {
	return a.sessionService.Revoke(sessionID)
}

func generateTokenHash() string {
	return uuid.New().String()
}
