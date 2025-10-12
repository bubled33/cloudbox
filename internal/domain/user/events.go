package user

import "github.com/google/uuid"

// Функции для генерации событий без времени

func NewUserCreatedEvent(u *User) (string, map[string]interface{}) {
	return "UserCreated", map[string]interface{}{
		"user_id":      u.ID,
		"email":        u.Email.String(),
		"display_name": u.DisplayName.String(),
	}
}

func NewUserDeletedEvent(userID uuid.UUID) (string, map[string]interface{}) {
	return "UserDeleted", map[string]interface{}{
		"user_id": userID,
	}
}

func NewUserEmailVerifiedEvent(userID uuid.UUID) (string, map[string]interface{}) {
	return "UserEmailVerified", map[string]interface{}{
		"user_id": userID,
	}
}
