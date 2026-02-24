package user

type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
}

type PhotoRequest struct {
	ID      *string `json:"id"`
	URL     string  `json:"url" binding:"required"`
	IsMain  bool    `json:"is_main"`
	Destroy *bool   `json:"_destroy"`
}

type UpdateProfileRequest struct {
	FullName        *string         `json:"full_name"`
	DateOfBirth     *string         `json:"date_of_birth"` // YYYY-MM-DD
	Status          *string         `json:"status"`        // onboarding, active
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
	Photos          *[]PhotoRequest `json:"photos"` // List of photo objects
}
