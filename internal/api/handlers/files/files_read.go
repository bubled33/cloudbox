package files_handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
)

// ListFiles godoc
// @Summary List user's files
// @Description Get paginated list of files with optional name search
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param q query string false "Search query by file name"
// @Param limit query int false "Results per page" default(20)
// @Param skip query int false "Number of results to skip" default(0)
// @Success 200 {object} ListFilesResponse "List of files"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to list files"
// @Router /files [get]
func (h *FileHandler) ListFiles(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := ctx.DefaultQuery("q", "")
	limitStr := ctx.DefaultQuery("limit", "20")
	skipStr := ctx.DefaultQuery("skip", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	skip, err := strconv.Atoi(skipStr)
	if err != nil || skip < 0 {
		skip = 0
	}

	files, total, err := h.fileService.SearchByName(ctx, userID, query, limit, skip)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	items := PresentFiles(files)

	fileResponses := make([]*FileResponse, 0, len(items))
	for i := range items {
		item := items[i]
		fileResponses = append(fileResponses, &item)
	}

	ctx.JSON(http.StatusOK, ListFilesResponse{
		Files: fileResponses,
		Total: total,
		Limit: limit,
		Skip:  skip,
	})
}

// GetFile godoc
// @Summary Get file detailed metadata
// @Description Get detailed metadata for a specific file including version count and ownership info
// @Tags files
// @Security Bearer
// @Accept json
// @Produce json
// @Param file_id path string true "File ID" format(uuid)
// @Success 200 {object} FileDetailResponse "File detailed metadata"
// @Failure 400 {object} map[string]string "Invalid file_id format"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to get file"
// @Router /files/{file_id} [get]
func (h *FileHandler) GetFile(ctx *gin.Context) {
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

	versions, err := h.fileVersionService.GetVersionsByFileID(ctx, fileID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	resp := PresentFileDetail(f, len(versions))

	ctx.JSON(http.StatusOK, resp)
}
