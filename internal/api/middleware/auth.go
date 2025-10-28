package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
)

func AuthMiddleware(authService *auth_service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		accessToken := parts[1]
		fmt.Println("Access token from header:", accessToken)

		session, err := authService.ValidateSessionByAccessToken(ctx, accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		ctx.Set("user_id", session.UserID)
		ctx.Set("session_id", session.ID)

		ctx.Next()
	}
}

func GetUserID(ctx *gin.Context) (uuid.UUID, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}
	return userID.(uuid.UUID), nil
}

func GetSessionID(ctx *gin.Context) (uuid.UUID, error) {
	sessionID, exists := ctx.Get("session_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("session_id not found in context")
	}
	return sessionID.(uuid.UUID), nil
}
