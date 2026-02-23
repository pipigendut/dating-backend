package usecases

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type UpdateProfileRequest struct {
	FullName        *string
	DateOfBirth     *string // YYYY-MM-DD
	Status          *entities.UserStatus
	Gender          *string
	HeightCM        *int
	Bio             *string
	InterestedIn    *string
	LookingFor      *string
	LocationCity    *string
	LocationCountry *string
	Interests       *[]string
	Languages       *[]string
}

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) GetProfile(id string) (*entities.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.NewBadRequest("invalid user id format")
	}

	user, err := u.repo.GetWithProfile(uid)
	if err != nil {
		return nil, errors.NewNotFound("user not found")
	}

	return user, nil
}
func (u *UserUsecase) UpdateProfile(userID uuid.UUID, data UpdateProfileRequest) error {
	user, err := u.repo.GetWithProfile(userID)
	if err != nil {
		return errors.NewNotFound("user not found")
	}

	if data.Status != nil {
		user.Status = *data.Status
	}

	if user.Profile == nil {
		user.Profile = &entities.Profile{UserID: userID}
	}

	if data.FullName != nil {
		user.Profile.FullName = *data.FullName
	}
	if data.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *data.DateOfBirth)
		if err != nil {
			return errors.NewBadRequest("invalid date format, use YYYY-MM-DD")
		}
		user.Profile.DateOfBirth = dob
	}
	if data.Gender != nil {
		user.Profile.Gender = *data.Gender
	}
	if data.HeightCM != nil {
		user.Profile.HeightCM = *data.HeightCM
	}
	if data.Bio != nil {
		user.Profile.Bio = *data.Bio
	}
	if data.InterestedIn != nil {
		user.Profile.InterestedIn = *data.InterestedIn
	}
	if data.LookingFor != nil {
		user.Profile.LookingFor = *data.LookingFor
	}
	if data.LocationCity != nil {
		user.Profile.LocationCity = *data.LocationCity
	}
	if data.LocationCountry != nil {
		user.Profile.LocationCountry = *data.LocationCountry
	}
	if data.Interests != nil {
		// Join slices into string or handle as needed
		user.Profile.Interests = "" // simplistic for now
	}
	if data.Languages != nil {
		user.Profile.Languages = "" // simplistic for now
	}

	return u.repo.Update(user)
}
