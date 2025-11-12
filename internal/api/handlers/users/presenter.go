package users_handler

import (
	"time"

	domainUser "github.com/yourusername/cloud-file-storage/internal/domain/user"
)

const timeFmt = time.RFC3339

func PresentUser(u *domainUser.User) GetMeResponse {
	return GetMeResponse{
		ID:              u.ID.String(),
		Email:           u.Email.String(),
		DisplayName:     u.DisplayName.String(),
		IsEmailVerified: u.IsEmailVerified,
		CreatedAt:       u.CreatedAt.UTC().Format(timeFmt),
	}
}

func PresentUpdateProfile(u *domainUser.User) UpdateProfileResponse {
	return UpdateProfileResponse{
		ID:              u.ID.String(),
		Email:           u.Email.String(),
		DisplayName:     u.DisplayName.String(),
		IsEmailVerified: u.IsEmailVerified,
		UpdatedAt:       u.UpdatedAt.UTC().Format(timeFmt),
	}
}
