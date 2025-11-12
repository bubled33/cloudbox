package files_handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

// DeletePublicLink godoc
// @Summary Delete public link
// @Description Revoke access to a public link
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Param link_id path string true "Public Link ID" format(uuid)
// @Success 200 {object} map[string]string "Link deleted"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File or link not found"
// @Failure 500 {object} map[string]string "Failed to delete public link"
// @Router /files/{file_id}/public-links/{link_id} [delete]
func (h *FileHandler) DeletePublicLink(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileID, err := uuid.Parse(ctx.Param("file_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid file_id format"})
		return
	}

	linkID, err := uuid.Parse(ctx.Param("link_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid link_id format"})
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

	if err := h.publicLinkService.Delete(ctx, linkID); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Public link deleted successfully"})
}

// GetPublicLinks godoc
// @Summary Get public links for file
// @Description Get list of all public links for a file
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Success 200 {object} ListPublicLinksResponse "List of public links"
// @Failure 400 {object} map[string]string "Invalid file_id format"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to get public links"
// @Router /files/{file_id}/public-links [get]
func (h *FileHandler) GetPublicLinks(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileID, err := uuid.Parse(ctx.Param("file_id"))
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

	links, err := h.publicLinkService.GetByFileID(ctx, fileID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	linkResponses := PresentPublicLinks(links)

	ctx.JSON(http.StatusOK, ListPublicLinksResponse{
		Links: linkResponses,
		Total: len(linkResponses),
	})
}

// CreatePublicLink godoc
// @Summary Create public link for file
// @Description Generate a public shareable link for downloading the file
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Param request body PublicLinkInput true "Expiration time"
// @Success 201 {object} PublicLinkResponse "Public link created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to create public link"
// @Router /public-links/{file_id} [post]
func (h *FileHandler) CreatePublicLink(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileID, err := uuid.Parse(ctx.Param("file_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid file_id format"})
		return
	}

	var input PublicLinkInput
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

	expiresAt := time.Now().Add(15 * time.Minute)
	if input.ExpiresIn != "" {
		dur, err := time.ParseDuration(input.ExpiresIn)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_in format (use 15m, 1h, 2h, etc)"})
			return
		}
		expiresAt = time.Now().Add(dur)
	}

	token := uuid.New().String()[:16]

	if f.Status.Equal(file_version.FileStatusProcessing) {
		ctx.Error(file_version.ErrVersionProcessing)
		return
	}

	if f.Status.Equal(file_version.FileStatusFailed) {
		ctx.Error(file_version.ErrVersionFailed)
		return
	}

	link, err := h.publicLinkService.Create(ctx, fileID, userID, token, expiresAt)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	resp := PresentPublicLink(link)

	ctx.JSON(http.StatusCreated, resp)
}

// DownloadByPublicLink godoc
// @Summary Download file by public link
// @Description Get presigned download URL for file using public link token (no authentication required)
// @Tags public
// @Accept json
// @Produce json
// @Param token path string true "Public link token"
// @Success 200 {object} PublicDownloadResponse "Download URL generated"
// @Failure 400 {object} map[string]string "Invalid token format"
// @Failure 404 {object} map[string]string "Public link not found or expired"
// @Failure 500 {object} map[string]string "Failed to generate download URL"
// @Router /public-links/{token} [get]
func (h *FileHandler) DownloadByPublicLink(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	link, err := h.publicLinkService.GetByToken(ctx, token)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if link == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "public link not found"})
		return
	}

	if link.IsExpired() {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "public link has expired"})
		return
	}

	file, err := h.fileService.GetByID(ctx, link.FileID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	versions, err := h.fileVersionService.GetVersionsByFileID(ctx, file.ID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var currentVersion *file_version.FileVersion
	for _, v := range versions {
		if v.VersionNum.Int() == file.VersionNum.Int() {
			currentVersion = v
			break
		}
	}
	if currentVersion == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file version not found"})
		return
	}

	downloadURL, err := h.fileVersionService.GetDownloadURL(ctx, currentVersion.ID, 1*time.Hour)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, PresentPublicDownload(file, currentVersion, *downloadURL))
}
