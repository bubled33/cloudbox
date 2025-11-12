package files_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
)

// UploadNewVersion godoc
// @Summary Upload new version of existing file
// @Description Create a new version of an existing file and get presigned URL for uploading file data
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Param request body UploadNewVersionInput true "New version metadata"
// @Success 201 {object} UploadFileResponse "New version created with upload URL"
// @Failure 400 {object} map[string]string "Invalid input or file size exceeds limit"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to create new version"
// @Router /files/{file_id}/versions [post]
func (h *FileHandler) UploadNewVersion(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionIDValue, exists := ctx.Get("session_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
		return
	}
	sessionID, ok := sessionIDValue.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid session id format"})
		return
	}

	fileIDStr := ctx.Param("file_id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid file_id format"})
		return
	}

	var input UploadNewVersionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := h.fileService.GetByID(ctx, fileID)
	if err != nil {
		_ = ctx.Error(err)
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

	const maxFileSize = 5 * 1024 * 1024 * 1024
	if input.Size > maxFileSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds maximum allowed (5GB)"})
		return
	}

	newVersionNum := file.VersionNum.Int() + 1
	updatedFile, version, uploadURL, err := h.fileVersionService.UploadNewVersion(
		ctx,
		fileID,
		userID,
		sessionID,
		input.Name,
		input.Size,
		input.Mime,
		newVersionNum,
	)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, UploadFileResponse{
		FileID:     updatedFile.ID.String(),
		VersionID:  version.ID.String(),
		UploadURL:  uploadURL,
		VersionNum: version.VersionNum.Int(),
		Status:     version.Status.String(),
		ExpiresIn:  "15m",
	})
}

// DeleteFile godoc
// @Summary Delete file
// @Description Delete a file and all its versions
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Success 200 {object} DeleteFileResponse "File deleted"
// @Failure 400 {object} map[string]string "Invalid file_id format"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to delete file"
// @Router /files/{file_id} [delete]
func (h *FileHandler) DeleteFile(ctx *gin.Context) {
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
		_ = ctx.Error(err)
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

	if err := h.fileService.Delete(ctx, fileID); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, DeleteFileResponse{
		Message: "File deleted successfully",
	})
}

// UpdateFile godoc
// @Summary Update file metadata
// @Description Update file name and metadata
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Param request body UpdateFileInput true "Update file data"
// @Success 200 {object} FileResponse "File updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to update file"
// @Router /files/{file_id} [patch]
func (h *FileHandler) UpdateFile(ctx *gin.Context) {
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

	var input UpdateFileInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	f, err := h.fileService.GetByID(ctx, fileID)
	if err != nil {
		_ = ctx.Error(err)
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

	if err := h.fileService.RenameFile(ctx, fileID, input.Name); err != nil {
		_ = ctx.Error(err)
		return
	}

	updatedFile, err := h.fileService.GetByID(ctx, fileID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	resp := PresentFile(updatedFile)
	ctx.JSON(http.StatusOK, resp)
}

// UploadNewFile godoc
// @Summary Create new file upload
// @Description Create a new file and get presigned URL for uploading file data
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body UploadFileInput true "File metadata"
// @Success 201 {object} UploadFileResponse "File created with upload URL"
// @Failure 400 {object} map[string]string "Invalid input or file size exceeds limit"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to create file"
// @Router /files [post]
func (h *FileHandler) UploadNewFile(ctx *gin.Context) {
	var input UploadFileInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownerIDValue, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	ownerID, ok := ownerIDValue.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	sessionIDValue, exists := ctx.Get("session_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
		return
	}
	sessionID, ok := sessionIDValue.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid session id format"})
		return
	}

	const maxFileSize = 5 * 1024 * 1024 * 1024
	if input.Size > maxFileSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds maximum allowed (5GB)"})
		return
	}

	file, version, uploadURL, err := h.fileVersionService.UploadNewFile(
		ctx,
		ownerID,
		sessionID,
		input.Name,
		input.Size,
		input.Mime,
	)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, UploadFileResponse{
		FileID:     file.ID.String(),
		VersionID:  version.ID.String(),
		UploadURL:  uploadURL,
		VersionNum: version.VersionNum.Int(),
		Status:     version.Status.String(),
		ExpiresIn:  "15m",
	})
}
