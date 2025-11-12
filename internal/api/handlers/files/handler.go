package files_handler

import (
	file_service "github.com/yourusername/cloud-file-storage/internal/app/file"
	file_version_service "github.com/yourusername/cloud-file-storage/internal/app/file_version"
	public_link_service "github.com/yourusername/cloud-file-storage/internal/app/public_link"
)

type FileHandler struct {
	fileVersionService *file_version_service.FileVersionService
	publicLinkService  *public_link_service.PublicLinkService
	fileService        *file_service.FileService
}

func NewFileHandler(
	fileVersionService *file_version_service.FileVersionService,
	fileService *file_service.FileService,
	publicLinkService *public_link_service.PublicLinkService,
) *FileHandler {
	return &FileHandler{
		fileVersionService: fileVersionService,
		fileService:        fileService,
		publicLinkService:  publicLinkService,
	}
}
