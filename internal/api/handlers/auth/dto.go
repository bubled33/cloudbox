package auth_handlers

// swagger:model
type RequestMagicLinkRequest struct {
	Email string `json:"email"`

	DisplayName *string `json:"display_name,omitempty"`
}

type RequestMagicLinkResponse struct {
	Message string `json:"message" example:"magic link sent to email"`
}

type VerifyMagicLinkResponse struct {
	Message      string `json:"message" example:"successfully authenticated"`
	SessionID    string `json:"session_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOi..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOi..."`
	ExpiresAt    string `json:"expires_at" example:"2025-11-04T12:00:00Z"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	Message      string `json:"message" example:"tokens refreshed"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOi..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOi..."`
	ExpiresAt    string `json:"expires_at" example:"2025-11-06T12:00:00Z"`
}

type LogoutResponse struct {
	Message string `json:"message" example:"successfully logged out"`
}

type ActiveSessionsResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

type SessionInfo struct {
	SessionID  string `json:"session_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	DeviceInfo string `json:"device_info" example:"Mozilla/5.0..."`
	IP         string `json:"ip" example:"192.168.1.1"`
	LastUsedAt string `json:"last_used_at" example:"2025-11-06T10:30:00Z"`
	CreatedAt  string `json:"created_at" example:"2025-11-05T08:00:00Z"`
	ExpiresAt  string `json:"expires_at" example:"2025-11-12T08:00:00Z"`
	IsCurrent  bool   `json:"is_current" example:"true"`
}
