package notification

import (
	"context"

	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type MailSender interface {
	SendMagicLink(ctx context.Context, tokenHash value_objects.TokenHash, to string, baseURL string) error
}
