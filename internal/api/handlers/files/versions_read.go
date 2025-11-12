package files_handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
)

// GetFileVersions godoc
// @Summary Get file versions
// @Description Get list of all versions for a file
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Success 200 {object} ListVersionsResponse "List of versions"
// @Failure 400 {object} map[string]string "Invalid file_id format"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to get versions"
// @Router /files/{file_id}/versions [get]
func (h *FileHandler) GetFileVersions(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileIDStr := ctx.Param("file_id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid file_id format"})
		return
	}

	f, err := h.fileService.GetByID(ctx, fileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file"})
		return
	}
	if f == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if f.OwnerID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	versions, err := h.fileVersionService.GetVersionsByFileID(ctx, fileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get versions"})
		return
	}

	versionResponses := make([]*FileVersionResponse, len(versions))
	for i, v := range versions {
		var previewKey *string
		if v.PreviewS3Key != nil {
			key := v.PreviewS3Key.String()
			previewKey = &key
		}
		versionResponses[i] = &FileVersionResponse{
			ID:           v.ID.String(),
			FileID:       v.FileId.String(),
			VersionNum:   v.VersionNum.Int(),
			Status:       v.Status.String(),
			Size:         v.Size.Uint64(),
			Mime:         v.Mime.String(),
			PreviewS3Key: previewKey,
			CreatedAt:    v.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    v.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}

	ctx.JSON(http.StatusOK, ListVersionsResponse{
		Versions: versionResponses,
		Total:    len(versionResponses),
	})
}
