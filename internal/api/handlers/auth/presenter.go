package auth_handlers

import (
	"time"

	domainSession "github.com/yourusername/cloud-file-storage/internal/domain/session"
)

const timeFmt = time.RFC3339

func PresentRequestMagicLinkOK() RequestMagicLinkResponse {
	return RequestMagicLinkResponse{Message: "magic link sent to email"}
}

func PresentVerify(session *domainSession.Session) VerifyMagicLinkResponse {
	return VerifyMagicLinkResponse{
		Message:      "successfully authenticated",
		SessionID:    session.ID.String(),
		AccessToken:  session.TokenHash.String(),
		RefreshToken: session.RefreshTokenHash.String(),
		ExpiresAt:    session.ExpiresAt.Time().UTC().Format(timeFmt),
	}
}
func PresentRefresh(session *domainSession.Session) RefreshTokenResponse {
	return RefreshTokenResponse{
		Message:      "tokens refreshed",
		AccessToken:  session.TokenHash.String(),
		RefreshToken: session.RefreshTokenHash.String(),
		ExpiresAt:    session.ExpiresAt.Time().UTC().Format(timeFmt),
	}
}

func PresentLogoutOK() LogoutResponse {
	return LogoutResponse{Message: "successfully logged out"}
}

func PresentLogoutAllOK() LogoutResponse {
	return LogoutResponse{Message: "all sessions revoked"}
}

func PresentRevokeSessionOK() LogoutResponse {
	return LogoutResponse{Message: "session revoked"}
}

func PresentActiveSessions(sessions []*domainSession.Session) ActiveSessionsResponse {
	sessionInfos := make([]SessionInfo, 0, len(sessions))

	for _, s := range sessions {
		sessionInfos = append(sessionInfos, SessionInfo{
			SessionID:  s.ID.String(),
			DeviceInfo: s.DeviceInfo.String(),
			IP:         s.Ip.String(),
			LastUsedAt: s.LastUsedAt.UTC().Format(timeFmt),
			CreatedAt:  s.CreatedAt.UTC().Format(timeFmt),
			ExpiresAt:  s.ExpiresAt.Time().UTC().Format(timeFmt),
			IsCurrent:  false,
		})
	}

	return ActiveSessionsResponse{Sessions: sessionInfos}
}
