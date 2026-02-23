package usecases

import (
	"context"
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
	Latitude        *float64
	Longitude       *float64
	Interests       *[]string
	Languages       *[]string
	Photos          *[]PhotoDTO
}

type UserUsecase struct {
	repo      repository.UserRepository
	storageUC *StorageUsecase
}

func NewUserUsecase(repo repository.UserRepository, storageUC *StorageUsecase) *UserUsecase {
	return &UserUsecase{repo: repo, storageUC: storageUC}
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
	if data.Latitude != nil {
		user.Profile.Latitude = data.Latitude
	}
	if data.Longitude != nil {
		user.Profile.Longitude = data.Longitude
	}
	if data.Interests != nil {
		user.Profile.Interests = ""
		for i, interest := range *data.Interests {
			if i > 0 {
				user.Profile.Interests += ","
			}
			user.Profile.Interests += interest
		}
	}
	if data.Languages != nil {
		user.Profile.Languages = ""
		for i, lang := range *data.Languages {
			if i > 0 {
				user.Profile.Languages += ","
			}
			user.Profile.Languages += lang
		}
	}

	if data.Photos != nil {
		// Replace all photos: Clear existing first
		// We use the repository through a transaction for safety if needed,
		// but GORM's Replace can also work if configured correctly.
		// For simplicity/reliability at this stage, we'll let GORM handle the sync via FullSaveAssociations
		// But we need to make sure the user.Photos slice is what we want the DB to reflect.
		newPhotos := make([]entities.Photo, 0)
		for i, p := range *data.Photos {
			newPhotos = append(newPhotos, entities.Photo{
				ID:        uuid.New(),
				UserID:    userID,
				URL:       p.URL,
				IsMain:    p.IsMain,
				SortOrder: i,
				CreatedAt: time.Now(),
			})
		}
		user.Photos = newPhotos
	}

	return u.repo.Update(user)
}

func (u *UserUsecase) DeleteAccount(userID uuid.UUID) error {
	// First fetch the user to get their photos
	user, err := u.GetProfile(userID.String())
	if err == nil && user != nil && u.storageUC != nil {
		// Attempt to delete all associated photos from storage concurrently or sequentially
		// We'll do it sequentially for simplicity
		for _, photo := range user.Photos {
			// In our schema, photo.URL holds the S3 key (e.g. users/UUID/profile/...)
			if photo.URL != "" {
				_ = u.storageUC.DeleteFile(context.Background(), photo.URL) // Fire and forget or handle gracefully
			}
		}
	}

	return u.repo.Delete(userID)
}
