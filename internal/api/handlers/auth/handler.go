package auth_handlers

import (
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
	session_service "github.com/yourusername/cloud-file-storage/internal/app/session"
)

type AuthHandler struct {
	authSrv    *auth_service.AuthService
	sessionSrv *session_service.SessionService // Добавлено
}

func NewAuthHandler(
	authSrv *auth_service.AuthService,
	sessionSrv *session_service.SessionService,
) *AuthHandler {
	return &AuthHandler{
		authSrv:    authSrv,
		sessionSrv: sessionSrv,
	}
}
