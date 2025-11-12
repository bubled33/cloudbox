package auth_handlers

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainSession "github.com/yourusername/cloud-file-storage/internal/domain/session"
)

// RequestMagicLink godoc
// @Summary Request magic link
// @Description Send magic link for authentication to user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RequestMagicLinkRequest true "User email and optional display name"
// @Success 200 {object} RequestMagicLinkResponse "Success message"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /magic-links [post]
func (h *AuthHandler) RequestMagicLink(ctx *gin.Context) {
	var req RequestMagicLinkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientIP := ctx.ClientIP()
	ip := net.ParseIP(clientIP)
	deviceInfo := ctx.GetHeader("User-Agent")

	if _, err := h.authSrv.RegisterWithMagicLink(ctx, req.Email, req.DisplayName, deviceInfo, ip); err != nil {

		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentRequestMagicLinkOK())
}

// VerifyMagicLink godoc
// @Summary Verify magic link
// @Description Verify magic link token and authenticate user
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Magic link token"
// @Success 200 {object} VerifyMagicLinkResponse "Session data with tokens"
// @Failure 400 {object} map[string]string "Missing token"
// @Failure 401 {object} map[string]string "Invalid or expired token"
// @Router /magic-links/{token} [get]
func (h *AuthHandler) VerifyMagicLink(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	clientIP := ctx.ClientIP()
	ip := net.ParseIP(clientIP)

	session, err := h.authSrv.Authenticate(ctx, token, ip)
	if err != nil {

		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentVerify(session))
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} RefreshTokenResponse "New tokens"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Router /tokens/refresh [post]
func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.sessionSrv.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentRefresh(session))
}

// Logout godoc
// @Summary Logout from current session
// @Description Revoke current session and invalidate tokens
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} LogoutResponse "Successfully logged out"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/sessions/current [delete]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	sessionID, exists := ctx.Get("session_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
		return
	}

	if err := h.sessionSrv.Revoke(ctx, sessionID.(uuid.UUID)); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentLogoutOK())
}

// LogoutAll godoc
// @Summary Logout from all sessions
// @Description Revoke all user sessions across all devices
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} LogoutResponse "All sessions revoked"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/sessions [delete]
func (h *AuthHandler) LogoutAll(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	if err := h.sessionSrv.RevokeAllForUser(ctx, userID.(uuid.UUID)); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentLogoutAllOK())
}

// GetActiveSessions godoc
// @Summary Get active sessions
// @Description Get list of all active sessions for current user
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ActiveSessionsResponse "List of active sessions"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	sessions, err := h.sessionSrv.GetByUserID(ctx, userID.(uuid.UUID))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	activeSessions := make([]*domainSession.Session, 0)
	for _, s := range sessions {
		if !s.IsRevoked && !s.IsExpired() {
			activeSessions = append(activeSessions, s)
		}
	}

	ctx.JSON(http.StatusOK, PresentActiveSessions(activeSessions))
}

// RevokeSession godoc
// @Summary Revoke specific session
// @Description Revoke specific session by ID
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} LogoutResponse "Session revoked"
// @Failure 400 {object} map[string]string "Invalid session ID"
// @Failure 403 {object} map[string]string "Not allowed to revoke this session"
// @Router /auth/sessions/{session_id} [delete]
func (h *AuthHandler) RevokeSession(ctx *gin.Context) {
	sessionIDStr := ctx.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	session, err := h.sessionSrv.GetByID(ctx, sessionID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	if session.UserID != currentUserID.(uuid.UUID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not allowed to revoke this session"})
		return
	}

	if err := h.sessionSrv.Revoke(ctx, sessionID); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentRevokeSessionOK())
}
