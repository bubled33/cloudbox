package session

import (
	"github.com/google/uuid"
)

// Сессия создана
func NewSessionCreatedEvent(sess *Session) (string, map[string]interface{}) {
	return "SessionCreated", map[string]interface{}{
		"session_id": sess.ID,
		"user_id":    sess.UserID,
		"expires_at": sess.ExpiresAt,
		"device":     sess.DeviceInfo,
		"ip":         sess.Ip.String(),
	}
}

// Сессия удалена
func NewSessionDeletedEvent(sessionID uuid.UUID) (string, map[string]interface{}) {
	return "SessionDeleted", map[string]interface{}{
		"session_id": sessionID,
	}
}

// Сессия отозвана
func NewSessionRevokedEvent(sessionID uuid.UUID) (string, map[string]interface{}) {
	return "SessionRevoked", map[string]interface{}{
		"session_id": sessionID,
	}
}

// Сессия истекла
func NewSessionExpiredEvent(sessionID uuid.UUID) (string, map[string]interface{}) {
	return "SessionExpired", map[string]interface{}{
		"session_id": sessionID,
	}
}
