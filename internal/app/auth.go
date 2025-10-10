package app

import (
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
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

// Запрос магической ссылки
func (a *AuthService) RequestMagicLink(userID uuid.UUID, deviceInfo string, ip net.IP) (*domain.MagicLink, error) {
	tokenHash := generateTokenHash()
	return a.magicLinkService.Create(userID, tokenHash, deviceInfo, "login", ip)
}

// Аутентификация по магической ссылке
func (a *AuthService) Authenticate(tokenHash string, ip net.IP) (*domain.Session, error) {
	link, err := a.magicLinkService.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, ErrInvalidMagicLink
	}

	if link.IsUsed || link.IsExpired {
		return nil, ErrInvalidMagicLink
	}

	// Маркируем ссылку как использованную
	if err := a.magicLinkService.MarkAsUsed(link.ID); err != nil {
		return nil, err
	}

	// Создаем сессию
	expiresAt := time.Now().Add(a.sessionTTL)
	session, err := a.sessionService.Create(
		link.UserID,
		generateTokenHash(),
		generateTokenHash(),
		link.DeviceInfo,
		ip,
		expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Валидация сессии
func (a *AuthService) ValidateSession(sessionID uuid.UUID) (*domain.Session, error) {
	session, err := a.sessionService.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsRevoked || session.IsExpired() {
		return nil, ErrInvalidSession
	}

	// Обновляем последнее использование через метод сервиса
	if err := a.sessionService.Touch(sessionID); err != nil {
		return nil, err
	}

	return session, nil
}

// Отзыв сессии
func (a *AuthService) RevokeSession(sessionID uuid.UUID) error {
	return a.sessionService.Revoke(sessionID)
}

// Вспомогательная функция генерации токена
func generateTokenHash() string {
	return uuid.New().String()
}
