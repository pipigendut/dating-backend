package auth

type CheckEmailRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type RegisterEmailRequest struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com"`
	Password    string `json:"password" binding:"required,min=8" example:"password123"`
	FullName    string `json:"full_name" binding:"required" example:"John Doe"`
	DateOfBirth string `json:"date_of_birth" binding:"required" example:"1995-01-01"`
}

type LoginEmailRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type GoogleLoginRequest struct {
	Email          string `json:"email" binding:"required,email" example:"user@example.com"`
	GoogleID       string `json:"google_id" binding:"required" example:"123456789"`
	FullName       string `json:"full_name" example:"John Doe"`
	ProfilePicture string `json:"profile_picture" example:"https://example.com/photo.jpg"`
}
