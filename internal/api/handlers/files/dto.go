package files_handler

type UploadFileInput struct {
	Name string `json:"name" binding:"required"`
	Size uint64 `json:"size" binding:"required,gt=0"`
	Mime string `json:"mime" binding:"required"`
}

type UploadFileResponse struct {
	FileID     string `json:"file_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	VersionID  string `json:"version_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	UploadURL  string `json:"upload_url" example:"https://s3.example.com/upload?..."`
	VersionNum int    `json:"version_num" example:"1"`
	Status     string `json:"status" example:"processing"`
	ExpiresIn  string `json:"expires_in" example:"15m"`
}

type FileResponse struct {
	ID           string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name         string  `json:"name" example:"document.pdf"`
	Mime         string  `json:"mime" example:"application/pdf"`
	Size         uint64  `json:"size" example:"1024000"`
	Status       string  `json:"status" example:"ready"`
	VersionNum   int     `json:"version_num" example:"1"`
	PreviewS3Key *string `json:"preview_s3_key" example:"files/user-id/file-id/preview.jpg"`
	CreatedAt    string  `json:"created_at" example:"2025-11-04T12:00:00Z"`
	UpdatedAt    string  `json:"updated_at" example:"2025-11-04T12:00:00Z"`
}

type FileDetailResponse struct {
	ID                string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	OwnerID           string  `json:"owner_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name              string  `json:"name" example:"document.pdf"`
	Mime              string  `json:"mime" example:"application/pdf"`
	Size              uint64  `json:"size" example:"1024000"`
	Status            string  `json:"status" example:"ready"`
	CurrentVersion    int     `json:"current_version" example:"3"`
	TotalVersions     int     `json:"total_versions" example:"5"`
	PreviewS3Key      *string `json:"preview_s3_key" example:"files/user-id/file-id/preview.jpg"`
	CreatedAt         string  `json:"created_at" example:"2025-11-04T12:00:00Z"`
	UpdatedAt         string  `json:"updated_at" example:"2025-11-04T12:00:00Z"`
	UploadedBySession string  `json:"uploaded_by_session_id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type ListFilesResponse struct {
	Files []*FileResponse `json:"files"`
	Total int64           `json:"total" example:"25"`
	Limit int             `json:"limit" example:"20"`
	Skip  int             `json:"skip" example:"0"`
}

type GetDownloadURLResponse struct {
	DownloadURL string `json:"download_url" example:"https://s3.example.com/download?..."`
	ExpiresIn   string `json:"expires_in" example:"1h"`
}

type UploadNewVersionInput struct {
	Name string `json:"name" binding:"required"`
	Size uint64 `json:"size" binding:"required,gt=0"`
	Mime string `json:"mime" binding:"required"`
}

type UpdateFileInput struct {
	Name string `json:"name" binding:"required,min=1"`
}

type FileVersionResponse struct {
	ID           string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174001"`
	FileID       string  `json:"file_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	VersionNum   int     `json:"version_num" example:"2"`
	Status       string  `json:"status" example:"ready"`
	Size         uint64  `json:"size" example:"2048000"`
	Mime         string  `json:"mime" example:"application/pdf"`
	PreviewS3Key *string `json:"preview_s3_key" example:"files/user-id/file-id/v2/preview.jpg"`
	CreatedAt    string  `json:"created_at" example:"2025-11-04T12:00:00Z"`
	UpdatedAt    string  `json:"updated_at" example:"2025-11-04T12:30:00Z"`
}

type ListVersionsResponse struct {
	Versions []*FileVersionResponse `json:"versions"`
	Total    int                    `json:"total" example:"5"`
}

type PublicLinkInput struct {
	ExpiresIn string `json:"expires_in" binding:"required" example:"24h"` // 15m, 1h, 24h, 7d, etc
}

type PublicLinkResponse struct {
	ID        string `json:"id" example:"123e4567-e89b-12d3-a456-426614174002"`
	FileID    string `json:"file_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Token     string `json:"token" example:"abc123def456"`
	IsExpired bool   `json:"is_expired" example:"false"`
	ExpiresAt string `json:"expires_at" example:"2025-11-05T12:00:00Z"`
	CreatedAt string `json:"created_at" example:"2025-11-04T12:00:00Z"`
}

type ListPublicLinksResponse struct {
	Links []*PublicLinkResponse `json:"links"`
	Total int                   `json:"total" example:"3"`
}

type PublicDownloadResponse struct {
	DownloadURL string `json:"download_url" example:"https://s3.amazonaws.com/...?X-Amz-Signature=..."`
	ExpiresIn   string `json:"expires_in" example:"1h"`
	FileName    string `json:"file_name" example:"document.pdf"`
	FileSize    uint64 `json:"file_size" example:"1024000"`
	Mime        string `json:"mime" example:"application/pdf"`
}

type DeleteFileResponse struct {
	Message string `json:"message" example:"File deleted successfully"`
}

type DeleteVersionResponse struct {
	Message string `json:"message" example:"Version deleted successfully"`
}
