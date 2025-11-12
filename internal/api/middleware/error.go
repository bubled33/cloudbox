package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/public_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

type apiError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err

		if c.Err() != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, apiError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input",
				Details: err.Error(),
			})
			return
		}

		status, ae := mapError(err)
		c.AbortWithStatusJSON(status, ae)
	}
}

func mapError(err error) (int, apiError) {
	switch {

	case errors.Is(err, domainerrors.ErrInvaliTokenHash):
		return http.StatusBadRequest, apiError{Code: "INVALID_TOKEN_HASH", Message: "Invalid token hash"}
	case errors.Is(err, domainerrors.ErrInvalidExpiry):
		return http.StatusBadRequest, apiError{Code: "INVALID_EXPIRES_AT", Message: "expiresAt must be in the future"}
	case errors.Is(err, domainerrors.ErrInvalidDeviceInfo):
		return http.StatusBadRequest, apiError{Code: "INVALID_DEVICE_INFO", Message: "Invalid device info"}
	case errors.Is(err, domainerrors.ErrInvalidIP):
		return http.StatusBadRequest, apiError{Code: "INVALID_IP", Message: "Invalid IP"}
	case errors.Is(err, domainerrors.ErrTransactionNotFound):
		return http.StatusNotFound, apiError{Code: "TRANSACTION_NOT_FOUND", Message: "Transaction not found"}

	case errors.Is(err, user.ErrNotFound):
		return http.StatusNotFound, apiError{Code: "USER_NOT_FOUND", Message: "User not found"}
	case errors.Is(err, user.ErrInvalidEmailFormat):
		return http.StatusBadRequest, apiError{Code: "INVALID_EMAIL_FORMAT", Message: "Invalid email format"}
	case errors.Is(err, user.ErrInvalidDisplayNameSize):
		return http.StatusBadRequest, apiError{Code: "INVALID_DISPLAY_NAME_SIZE", Message: "Display name must be between 2 and 50 characters"}

	case errors.Is(err, session.ErrInvaliTokenHash):
		return http.StatusBadRequest, apiError{Code: "INVALID_TOKEN_HASH", Message: "Invalid token hash"}
	case errors.Is(err, session.ErrNotFound):
		return http.StatusNotFound, apiError{Code: "SESSION_NOT_FOUND", Message: "Session not found"}
	case errors.Is(err, session.ErrInvalidSession):
		return http.StatusBadRequest, apiError{Code: "INVALID_SESSION", Message: "Invalid session"}
	case errors.Is(err, session.ErrInvalidExpiry):
		return http.StatusBadRequest, apiError{Code: "INVALID_EXPIRES_AT", Message: "expiresAt must be in the future"}
	case errors.Is(err, session.ErrInvalidDeviceInfo):
		return http.StatusBadRequest, apiError{Code: "INVALID_DEVICE_INFO", Message: "Invalid device info"}
	case errors.Is(err, session.ErrInvalidIP):
		return http.StatusBadRequest, apiError{Code: "INVALID_IP", Message: "Invalid IP"}

	case errors.Is(err, public_link.ErrNotFound):
		return http.StatusNotFound, apiError{Code: "PUBLIC_LINK_NOT_FOUND", Message: "Public link not found"}
	case errors.Is(err, public_link.ErrInvalidExpiryTime):
		return http.StatusBadRequest, apiError{Code: "INVALID_EXPIRES_AT", Message: "expiresAt must be in the future"}

	case errors.Is(err, magic_link.ErrNotFound):
		return http.StatusNotFound, apiError{Code: "MAGIC_LINK_NOT_FOUND", Message: "Magic link not found"}
	case errors.Is(err, magic_link.ErrInvalid):
		return http.StatusBadRequest, apiError{Code: "INVALID_MAGIC_LINK", Message: "Magic link is invalid"}
	case errors.Is(err, magic_link.ErrMagicLink):

		return http.StatusBadRequest, apiError{Code: "MAGIC_LINK_INVALID_OR_EXPIRED", Message: "Invalid or expired magic link"}

	case errors.Is(err, file_version.ErrVersionNotFound):
		return http.StatusNotFound, apiError{Code: "FILE_VERSION_NOT_FOUND", Message: "Version not found"}
	case errors.Is(err, file_version.ErrCannotDeleteCurr):

		return http.StatusBadRequest, apiError{Code: "CANNOT_DELETE_CURRENT_VERSION", Message: "Cannot delete current version"}
	case errors.Is(err, file_version.ErrVersionProcessing):

		return http.StatusBadRequest, apiError{Code: "VERSION_PROCESSING", Message: "Cannot delete file, some versions are processing"}

	case errors.Is(err, file.ErrNotFound):
		return http.StatusNotFound, apiError{Code: "FILE_NOT_FOUND", Message: "File not found"}

	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	}
}
