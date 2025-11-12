package users_handler

type GetMeResponse struct {
	ID              string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email           string `json:"email" example:"user@example.com"`
	DisplayName     string `json:"display_name" example:"John Doe"`
	IsEmailVerified bool   `json:"is_email_verified" example:"true"`
	CreatedAt       string `json:"created_at" example:"2025-11-04T12:00:00Z"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=1,max=255"`
	Email       string `json:"email" binding:"required,email"`
}

type UpdateProfileResponse struct {
	ID              string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email           string `json:"email" example:"user@example.com"`
	DisplayName     string `json:"display_name" example:"John Doe"`
	IsEmailVerified bool   `json:"is_email_verified" example:"true"`
	UpdatedAt       string `json:"updated_at" example:"2025-11-04T12:00:00Z"`
}

type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" binding:"required,eq=DELETE_MY_ACCOUNT"`
}

type DeleteAccountResponse struct {
	Message string `json:"message" example:"Account successfully deleted"`
}
