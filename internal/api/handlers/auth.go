package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
)

type AuthHandler struct {
	auth_srv auth_service.AuthService
}

func NewAuthHandler(auth_srv auth_service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth_srv: auth_srv,
	}
}

func (h *AuthHandler) RequestMagicLink(ctx *gin.Context) {
	fmt.Println("Consume request!")
}
