package magic_link

// NewMagicLinkCreatedEvent формирует событие создания магической ссылки
func NewMagicLinkCreatedEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkCreated", map[string]interface{}{
		"link_id":    link.ID,
		"user_id":    link.UserID,
		"purpose":    link.Purpose.String(),
		"expires_at": link.ExpiredAt.Time(),
	}
}

// NewMagicLinkUsedEvent формирует событие использования магической ссылки
func NewMagicLinkUsedEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkUsed", map[string]interface{}{
		"link_id": link.ID,
		"user_id": link.UserID,
	}
}

// NewMagicLinkExpiredEvent формирует событие истечения срока действия магической ссылки
func NewMagicLinkExpiredEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkExpired", map[string]interface{}{
		"link_id": link.ID,
		"user_id": link.UserID,
	}
}

// NewMagicLinkDeletedEvent формирует событие удаления магической ссылки
func NewMagicLinkDeletedEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkDeleted", map[string]interface{}{
		"link_id": link.ID,
		"user_id": link.UserID,
	}
}
