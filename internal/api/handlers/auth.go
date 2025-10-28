package handlers

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
)

type AuthHandler struct {
	Auth_srv *auth_service.AuthService
}

func NewAuthHandler(auth_srv *auth_service.AuthService) *AuthHandler {
	return &AuthHandler{
		Auth_srv: auth_srv,
	}
}

type UserInput struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func (h *AuthHandler) RequestMagicLink(ctx *gin.Context) {
	fmt.Println("Consume request!")
	var input UserInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientIP := ctx.ClientIP()
	ip := net.ParseIP(clientIP)

	deviceInfo := ctx.GetHeader("User-Agent")
	_, err := h.Auth_srv.RegisterWithMagicLink(ctx, input.Email, input.DisplayName, deviceInfo, ip)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func (h *AuthHandler) VerifyMagicLink(ctx *gin.Context) {
	fmt.Println("Consume request!")
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	clientIP := ctx.ClientIP()
	ip := net.ParseIP(clientIP)

	session, err := h.Auth_srv.Authenticate(ctx, token, ip)
	fmt.Println(err)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "successfully authenticated",
		"session_id":    session.ID,
		"access_token":  session.TokenHash.String(),
		"refresh_token": session.RefreshTokenHash.String(),
	})
}
