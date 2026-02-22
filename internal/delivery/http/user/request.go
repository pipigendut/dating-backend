package user

type UserUpdateRequest struct {
	Bio      *string `json:"bio" example:"Avid hiker and coffee lover."`
	FullName *string `json:"full_name" example:"John Doe"`
}
