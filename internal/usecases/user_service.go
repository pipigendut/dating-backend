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

	user, err := u.repo.GetWithRelations(uid)
	if err != nil {
		return nil, errors.NewNotFound("user not found")
	}

	for i := range user.Photos {
		if user.Photos[i].URL != "" && u.storageUC != nil {
			user.Photos[i].URL = u.storageUC.GetPublicURL(user.Photos[i].URL)
		}
	}

	return user, nil
}
func (u *UserUsecase) UpdateProfile(userID uuid.UUID, data UpdateProfileRequest) error {
	user, err := u.repo.GetWithRelations(userID)
	if err != nil {
		return errors.NewNotFound("user not found")
	}

	if data.Status != nil {
		user.Status = *data.Status
	}

	if data.Gender != nil {
		if *data.Gender == "" {
			user.GenderID = nil
		} else {
			uid, err := uuid.Parse(*data.Gender)
			if err != nil {
				return errors.NewBadRequest("invalid gender id format")
			}
			user.GenderID = &uid
		}
	}

	if data.LookingFor != nil {
		if *data.LookingFor == "" {
			user.RelationshipTypeID = nil
		} else {
			uid, err := uuid.Parse(*data.LookingFor)
			if err != nil {
				return errors.NewBadRequest("invalid relationship type id format")
			}
			user.RelationshipTypeID = &uid
		}
	}

	if data.InterestedIn != nil {
		if *data.InterestedIn == "" {
			user.InterestedGenders = nil
		} else {
			uid, err := uuid.Parse(*data.InterestedIn)
			if err == nil {
				user.InterestedGenders = []entities.MasterGender{{ID: uid}}
			}
		}
	}

	if data.FullName != nil {
		user.FullName = *data.FullName
	}
	if data.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *data.DateOfBirth)
		if err != nil {
			return errors.NewBadRequest("invalid date format, use YYYY-MM-DD")
		}
		user.DateOfBirth = dob
	}
	if data.HeightCM != nil {
		user.HeightCM = *data.HeightCM
	}
	if data.Bio != nil {
		user.Bio = *data.Bio
	}
	if data.LocationCity != nil {
		user.LocationCity = *data.LocationCity
	}
	if data.LocationCountry != nil {
		user.LocationCountry = *data.LocationCountry
	}
	if data.Latitude != nil {
		user.Latitude = data.Latitude
	}
	if data.Longitude != nil {
		user.Longitude = data.Longitude
	}

	if data.Interests != nil {
		var newInterests []entities.MasterInterest
		for _, rawID := range *data.Interests {
			uid, err := uuid.Parse(rawID)
			if err == nil {
				newInterests = append(newInterests, entities.MasterInterest{ID: uid})
			}
		}
		user.Interests = newInterests
	}

	if data.Languages != nil {
		var newLanguages []entities.MasterLanguage
		for _, rawID := range *data.Languages {
			uid, err := uuid.Parse(rawID)
			if err == nil {
				newLanguages = append(newLanguages, entities.MasterLanguage{ID: uid})
			}
		}
		user.Languages = newLanguages
	}

	if data.Photos != nil {
		newPhotos := make([]entities.Photo, 0)
		newPhotoIDs := make(map[uuid.UUID]bool)
		for i, p := range *data.Photos {
			pid := uuid.New()
			if p.ID != nil && *p.ID != "" {
				if parsed, err := uuid.Parse(*p.ID); err == nil {
					pid = parsed
				}
			}

			// If frontend says destroy, skip adding it to the newPhoto array and diff
			if p.Destroy != nil && *p.Destroy {
				continue
			}

			newPhotoIDs[pid] = true
			newPhotos = append(newPhotos, entities.Photo{
				ID:        pid,
				UserID:    userID,
				URL:       p.URL,
				IsMain:    p.IsMain,
				SortOrder: i,
				CreatedAt: time.Now(),
			})
		}

		// Find deleted photos to remove from S3 and DB
		existingUser, err := u.repo.GetWithRelations(userID)
		if err == nil && existingUser != nil {
			for _, existingPhoto := range existingUser.Photos {
				if !newPhotoIDs[existingPhoto.ID] {
					// Delete from S3
					if u.storageUC != nil && existingPhoto.URL != "" {
						_ = u.storageUC.DeleteFile(context.Background(), existingPhoto.URL)
					}
					// Explicitly delete from DB to avoid orphan records
					_ = u.repo.DeletePhoto(existingPhoto.ID)
				}
			}
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
