package users_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
	user_service "github.com/yourusername/cloud-file-storage/internal/app/user"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

type UserHandler struct {
	userSrv *user_service.UserService
}

func NewUserHandler(userSrv *user_service.UserService) *UserHandler {
	return &UserHandler{userSrv: userSrv}
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update user's display name and email
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update data"
// @Success 200 {object} UpdateProfileResponse "Profile updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Email already in use"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [patch]
func (h *UserHandler) UpdateProfile(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if req.Email != u.Email.String() {
		existing, err := h.userSrv.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		if err != nil && err != user.ErrNotFound {
			_ = ctx.Error(err)
			return
		}
	}

	updated, err := h.userSrv.UpdateProfile(ctx, userID, req.Email, req.DisplayName)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentUpdateProfile(updated))
}

// DeleteAccount godoc
// @Summary Delete user account
// @Description Permanently delete user account and all associated data
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body DeleteAccountRequest true "Confirmation required"
// @Success 200 {object} DeleteAccountResponse "Account deleted"
// @Failure 400 {object} map[string]string "Invalid confirmation"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [delete]
func (h *UserHandler) DeleteAccount(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req DeleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Confirmation != "DELETE_MY_ACCOUNT" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid confirmation"})
		return
	}

	if err := h.userSrv.DeleteAccount(ctx, userID); err != nil {
		if err == user.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, DeleteAccountResponse{
		Message: "Account successfully deleted",
	})
}
