package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
	user_service "github.com/yourusername/cloud-file-storage/internal/app/user"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

type UserHandler struct {
	user_srv *user_service.UserService
}

func NewUserHandler(user_srv *user_service.UserService) *UserHandler {
	return &UserHandler{
		user_srv: user_srv,
	}
}

func (h *UserHandler) GetMe(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	u, err := h.user_srv.GetByID(ctx, userID)
	if err != nil {
		if err == user.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		fmt.Printf("Error getting user %s: %v\n", userID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":                u.ID,
		"email":             u.Email.String(),
		"display_name":      u.DisplayName.String(),
		"is_email_verified": u.IsEmailVerified,
		"created_at":        u.CreatedAt,
	})
}
