package user

type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
}

type UpdateProfileRequest struct {
	FullName        *string   `json:"full_name"`
	DateOfBirth     *string   `json:"date_of_birth"` // YYYY-MM-DD
	Status          *string   `json:"status"`        // onboarding, active
	Gender          *string   `json:"gender"`
	HeightCM        *int      `json:"height_cm"`
	Bio             *string   `json:"bio"`
	InterestedIn    *string   `json:"interested_in"`
	LookingFor      *string   `json:"looking_for"`
	LocationCity    *string   `json:"location_city"`
	LocationCountry *string   `json:"location_country"`
	Interests       *[]string `json:"interests"`
	Languages       *[]string `json:"languages"`
}
