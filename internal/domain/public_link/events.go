package public_link

import (
	"github.com/google/uuid"
)

func NewPublicLinkCreatedEvent(link *PublicLink) (string, map[string]interface{}) {
	return "PublicLinkCreated", map[string]interface{}{
		"link_id":    link.ID,
		"file_id":    link.FileID,
		"created_by": link.CreatedByUserID,
		"expires_at": link.ExpiredAt,
	}
}

func NewPublicLinkDeletedEvent(linkID uuid.UUID) (string, map[string]interface{}) {
	return "PublicLinkDeleted", map[string]interface{}{
		"link_id": linkID,
	}
}

func NewPublicLinkExpiredEvent(linkID uuid.UUID) (string, map[string]interface{}) {
	return "PublicLinkExpired", map[string]interface{}{
		"link_id": linkID,
	}
}
