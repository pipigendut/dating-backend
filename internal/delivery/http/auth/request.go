package auth

type CheckEmailRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type PhotoRequest struct {
	URL    string `json:"url" binding:"required"`
	IsMain bool   `json:"is_main"`
}

type DeviceRequest struct {
	DeviceID    string `json:"device_id" binding:"required"`
	DeviceName  string `json:"device_name"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`
	AppVersion  string `json:"app_version"`
	FCMToken    string `json:"fcm_token"`
}

type RegisterEmailRequest struct {
	ID              *string         `json:"id"`
	Email           string          `json:"email" binding:"required,email" example:"user@example.com"`
	Password        string          `json:"password" binding:"required,min=8" example:"password123"`
	FullName        string          `json:"full_name" binding:"required" example:"John Doe"`
	DateOfBirth     string          `json:"date_of_birth" binding:"required" example:"1995-01-01"`
	Gender          *string         `json:"gender"`
	HeightCM        *int            `json:"height_cm"`
	Bio             *string         `json:"bio"`
	InterestedIn    *string         `json:"interested_in"`
	LookingFor      *string         `json:"looking_for"`
	LocationCity    *string         `json:"location_city"`
	LocationCountry *string         `json:"location_country"`
	Latitude        *float64        `json:"latitude"`
	Longitude       *float64        `json:"longitude"`
	Interests       *[]string       `json:"interests"`
	Languages       *[]string       `json:"languages"`
	Photos          *[]PhotoRequest `json:"photos"`
	Device          DeviceRequest   `json:"device"`
}

type LoginEmailRequest struct {
	Email    string        `json:"email" binding:"required,email" example:"user@example.com"`
	Password string        `json:"password" binding:"required" example:"password123"`
	Device   DeviceRequest `json:"device"`
}

type GoogleLoginRequest struct {
	ID              *string         `json:"id"`
	Email           string          `json:"email" binding:"required,email" example:"user@example.com"`
	GoogleID        string          `json:"google_id" binding:"required" example:"123456789"`
	FullName        string          `json:"full_name" example:"John Doe"`
	ProfilePicture  string          `json:"profile_picture" example:"https://example.com/photo.jpg"`
	DateOfBirth     *string         `json:"date_of_birth"`
	Gender          *string         `json:"gender"`
	HeightCM        *int            `json:"height_cm"`
	Bio             *string         `json:"bio"`
	InterestedIn    *string         `json:"interested_in"`
	LookingFor      *string         `json:"looking_for"`
	LocationCity    *string         `json:"location_city"`
	LocationCountry *string         `json:"location_country"`
	Latitude        *float64        `json:"latitude"`
	Longitude       *float64        `json:"longitude"`
	Interests       *[]string       `json:"interests"`
	Languages       *[]string       `json:"languages"`
	Photos          *[]PhotoRequest `json:"photos"`
	Device          DeviceRequest   `json:"device"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	DeviceID     string `json:"device_id" binding:"required"`
}

type LogoutRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}
