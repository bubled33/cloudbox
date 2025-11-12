package magic_link

func NewMagicLinkCreatedEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkCreated", map[string]interface{}{
		"link_id":    link.ID,
		"user_id":    link.UserID,
		"purpose":    link.Purpose.String(),
		"expires_at": link.ExpiredAt.Time(),
	}
}

func NewMagicLinkUsedEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkUsed", map[string]interface{}{
		"link_id": link.ID,
		"user_id": link.UserID,
	}
}

func NewMagicLinkExpiredEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkExpired", map[string]interface{}{
		"link_id": link.ID,
		"user_id": link.UserID,
	}
}

func NewMagicLinkDeletedEvent(link *MagicLink) (string, map[string]interface{}) {
	return "MagicLinkDeleted", map[string]interface{}{
		"link_id": link.ID,
		"user_id": link.UserID,
	}
}
