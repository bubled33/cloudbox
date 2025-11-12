package auth_service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	magic_link_service "github.com/yourusername/cloud-file-storage/internal/app/magic_link"
	session_service "github.com/yourusername/cloud-file-storage/internal/app/session"
	user_service "github.com/yourusername/cloud-file-storage/internal/app/user"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/notification"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type AuthService struct {
	magicLinkService *magic_link_service.MagicLinkService
	userService      *user_service.UserService
	sessionService   *session_service.SessionService
	mailSender       notification.MailSender
	sessionTTL       time.Duration
}

func NewAuthService(mlService *magic_link_service.MagicLinkService, sService *session_service.SessionService, uService *user_service.UserService, sessionTTL time.Duration, mailSender notification.MailSender) *AuthService {
	return &AuthService{
		magicLinkService: mlService,
		userService:      uService,
		sessionService:   sService,
		sessionTTL:       sessionTTL,
		mailSender:       mailSender,
	}
}

func (a *AuthService) RequestMagicLink(ctx context.Context, userID uuid.UUID, deviceInfo string, ip net.IP) (*magic_link.MagicLink, error) {
	tokenHash := generateTokenHash()
	m, err := a.magicLinkService.Create(ctx, userID, tokenHash, deviceInfo, "login", ip)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, fmt.Errorf("magic link creation returned nil")
	}
	fmt.Println(tokenHash, " was sended")
	tokenObj, err := value_objects.NewTokenHash(tokenHash)
	if err != nil {
		return nil, err
	}

	u, err := a.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	a.mailSender.SendMagicLink(ctx, tokenObj, u.Email.String(), "https://localhost:8030")
	fmt.Println(m.TokenHash.String())

	return m, nil
}

func (a *AuthService) RegisterWithMagicLink(ctx context.Context, email string, displayName *string, deviceInfo string, ip net.IP) (*magic_link.MagicLink, error) {

	u, err := a.userService.GetByEmail(ctx, email)
	fmt.Println("email 1", email)
	if err == user.ErrNotFound {

		if displayName == nil {
			return nil, errors.New("display name required")
		}
		u, err = a.userService.Create(ctx, email, *displayName)
		fmt.Print(err)
		if err != nil {
			return nil, err
		}
	}

	return a.RequestMagicLink(ctx, u.ID, deviceInfo, ip)
}

func (a *AuthService) Authenticate(ctx context.Context, tokenHash string, ip net.IP) (*session.Session, error) {
	link, err := a.magicLinkService.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, magic_link.ErrMagicLink
	}

	if !link.IsValid() {
		return nil, magic_link.ErrMagicLink
	}

	if err := a.magicLinkService.MarkAsUsed(ctx, link.ID); err != nil {
		return nil, err
	}

	if err := a.userService.VerifyEmail(ctx, link.UserID); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(a.sessionTTL)
	session, err := a.sessionService.Create(
		ctx,
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

func (a *AuthService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*session.Session, error) {
	sess, err := a.sessionService.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if sess.IsRevoked || sess.IsExpired() {
		return nil, session.ErrInvalidSession
	}

	if err := a.sessionService.Touch(ctx, sessionID); err != nil {
		return nil, err
	}

	return sess, nil
}

func (a *AuthService) ValidateSessionByAccessToken(ctx context.Context, accessToken string) (*session.Session, error) {

	sess, err := a.sessionService.GetByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	if sess.IsRevoked || sess.IsExpired() {
		return nil, session.ErrInvalidSession
	}

	if err := a.sessionService.Touch(ctx, sess.ID); err != nil {
		return nil, err
	}

	return sess, nil
}

func (a *AuthService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return a.sessionService.Revoke(ctx, sessionID)
}

func generateTokenHash() string {
	return uuid.New().String()
}
