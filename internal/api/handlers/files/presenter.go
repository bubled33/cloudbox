package files_handler

import (
	"time"

	domainFile "github.com/yourusername/cloud-file-storage/internal/domain/file"
	domainVer "github.com/yourusername/cloud-file-storage/internal/domain/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/public_link"
)

const timeFmt = time.RFC3339

func PresentFile(f *domainFile.File) FileResponse {
	var preview *string
	if f.PreviewS3Key != nil {
		k := f.PreviewS3Key.String()
		preview = &k
	}
	return FileResponse{
		ID:           f.ID.String(),
		Name:         f.Name.String(),
		Mime:         f.Mime.String(),
		Size:         f.Size.Uint64(),
		Status:       f.Status.String(),
		VersionNum:   f.VersionNum.Int(),
		PreviewS3Key: preview,
		CreatedAt:    f.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    f.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func PresentVersion(v *domainVer.FileVersion) FileVersionResponse {
	var preview *string
	if v.PreviewS3Key != nil {
		k := v.PreviewS3Key.String()
		preview = &k
	}
	return FileVersionResponse{
		ID:           v.ID.String(),
		FileID:       v.FileId.String(),
		VersionNum:   v.VersionNum.Int(),
		Status:       v.Status.String(),
		Size:         v.Size.Uint64(),
		Mime:         v.Mime.String(),
		PreviewS3Key: preview,
		CreatedAt:    v.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    v.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func PresentFiles(files []*domainFile.File) []FileResponse {
	out := make([]FileResponse, 0, len(files))
	for _, f := range files {
		if f == nil {
			continue
		}
		out = append(out, PresentFile(f))
	}
	return out
}

func PresentVersions(versions []*domainVer.FileVersion) []FileVersionResponse {
	out := make([]FileVersionResponse, 0, len(versions))
	for _, v := range versions {
		if v == nil {
			continue
		}
		out = append(out, PresentVersion(v))
	}
	return out
}

func PresentFileDetail(f *domainFile.File, totalVersions int) FileDetailResponse {
	var preview *string
	if f.PreviewS3Key != nil {
		k := f.PreviewS3Key.String()
		preview = &k
	}
	return FileDetailResponse{
		ID:                f.ID.String(),
		OwnerID:           f.OwnerID.String(),
		Name:              f.Name.String(),
		Mime:              f.Mime.String(),
		Size:              f.Size.Uint64(),
		Status:            f.Status.String(),
		CurrentVersion:    f.VersionNum.Int(),
		TotalVersions:     totalVersions,
		PreviewS3Key:      preview,
		CreatedAt:         f.CreatedAt.UTC().Format(timeFmt),
		UpdatedAt:         f.UpdatedAt.UTC().Format(timeFmt),
		UploadedBySession: f.UploadedBySessionId.String(),
	}
}

// PresentPublicLink преобразует доменную PublicLink в DTO
func PresentPublicLink(link *public_link.PublicLink) PublicLinkResponse {
	return PublicLinkResponse{
		ID:        link.ID.String(),
		FileID:    link.FileID.String(),
		Token:     link.TokenHash,
		ExpiresAt: link.ExpiredAt.UTC().Format(timeFmt),
		IsExpired: link.IsExpired(),
		CreatedAt: link.CreatedAt.UTC().Format(timeFmt),
	}
}

// PresentPublicLinks преобразует срез PublicLink в []PublicLinkResponse
func PresentPublicLinks(links []*public_link.PublicLink) []*PublicLinkResponse {
	out := make([]*PublicLinkResponse, 0, len(links))
	for _, l := range links {
		if l == nil {
			continue
		}
		resp := PresentPublicLink(l)
		out = append(out, &resp)
	}
	return out
}

func PresentPublicDownload(file *domainFile.File, version *domainVer.FileVersion, downloadURL string) PublicDownloadResponse {
	return PublicDownloadResponse{
		DownloadURL: downloadURL,
		ExpiresIn:   "1h",
		FileName:    file.Name.String(),
		FileSize:    version.Size.Uint64(),
		Mime:        version.Mime.String(),
	}
}
