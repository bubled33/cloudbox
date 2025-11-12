package users_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

// GetMe godoc
// @Summary Get current user profile
// @Description Retrieve authenticated user's profile information
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} GetMeResponse "User profile data"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) GetMe(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	u, err := h.userSrv.GetByID(ctx, userID)
	if err != nil {

		if err == user.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentUser(u))
}
