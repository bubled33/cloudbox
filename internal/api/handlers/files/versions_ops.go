package files_handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

// GetVersionDownloadURL godoc
// @Summary Get download URL for file version
// @Description Generate presigned URL for downloading specific version of file (expires in 1 hour). If version_num is omitted or 0, uses current file version.
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Param version_num path int false "Version number (optional, defaults to current version)" default(0)
// @Success 200 {object} GetDownloadURLResponse "Download URL generated"
// @Failure 400 {object} map[string]string "Invalid file_id format"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File or version not found"
// @Failure 500 {object} map[string]string "Failed to generate download URL"
// @Router /files/{file_id}/versions/{version_num}/content [get]
func (h *FileHandler) GetVersionDownloadURL(ctx *gin.Context) {
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

	file, err := h.fileService.GetByID(ctx, fileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file"})
		return
	}
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if file.OwnerID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	versionNumStr := ctx.Param("version_num")
	versionNum := 0
	if versionNumStr != "" {
		versionNum, err = strconv.Atoi(versionNumStr)
		if err != nil || versionNum < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid version_num format"})
			return
		}
	}

	var targetVersion *file_version.FileVersion

	if versionNum == 0 {
		versions, err := h.fileVersionService.GetVersionsByFileID(ctx, fileID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file versions"})
			return
		}

		for _, v := range versions {
			if v.VersionNum.Int() == file.VersionNum.Int() {
				targetVersion = v
				break
			}
		}
	} else {
		versions, err := h.fileVersionService.GetVersionsByFileID(ctx, fileID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file versions"})
			return
		}

		for _, v := range versions {
			if v.VersionNum.Int() == versionNum {
				targetVersion = v
				break
			}
		}
	}

	if targetVersion == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	downloadURL, err := h.fileVersionService.GetDownloadURL(ctx, targetVersion.ID, 1*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download url"})
		return
	}

	ctx.JSON(http.StatusOK, GetDownloadURLResponse{
		DownloadURL: *downloadURL,
		ExpiresIn:   "1h",
	})
}

// RestoreFileVersion godoc
// @Summary Restore file version
// @Description Make a specific version the current version
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Param version_num path int true "Version number to restore"
// @Success 200 {object} FileResponse "Version restored"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File or version not found"
// @Failure 500 {object} map[string]string "Failed to restore version"
// @Router /files/{file_id}/versions/{version_num}/restore [post]
func (h *FileHandler) RestoreFileVersion(ctx *gin.Context) {
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

	versionNumStr := ctx.Param("version_num")
	versionNum, err := strconv.Atoi(versionNumStr)
	if err != nil || versionNum <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid version_num format"})
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

	var versionToRestore *file_version.FileVersion
	for _, v := range versions {
		if v.VersionNum.Int() == versionNum {
			versionToRestore = v
			break
		}
	}

	if versionToRestore == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	err = h.fileVersionService.RestoreVersion(ctx, fileID, versionToRestore.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore version"})
		return
	}

	updatedFile, err := h.fileService.GetByID(ctx, fileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updated file"})
		return
	}

	var previewKey *string
	if updatedFile.PreviewS3Key != nil {
		key := updatedFile.PreviewS3Key.String()
		previewKey = &key
	}

	resp := FileResponse{
		ID:           updatedFile.ID.String(),
		Name:         updatedFile.Name.String(),
		Mime:         updatedFile.Mime.String(),
		Size:         updatedFile.Size.Uint64(),
		Status:       updatedFile.Status.String(),
		VersionNum:   updatedFile.VersionNum.Int(),
		PreviewS3Key: previewKey,
		CreatedAt:    updatedFile.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    updatedFile.UpdatedAt.UTC().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, resp)
}
