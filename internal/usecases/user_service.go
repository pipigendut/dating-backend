package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/background"
	"github.com/pipigendut/dating-backend/internal/background/jobs"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type WsDisconnecter interface {
	DisconnectUser(userID uuid.UUID)
}

type UpdateProfileRequest struct {
	FullName        *string
	DateOfBirth     *string // YYYY-MM-DD
	Status          *entities.UserStatus
	Gender          *string
	HeightCM        *int
	Bio             *string
	InterestedIn    *string
	RelationshipType *string
	LocationCity    *string
	LocationCountry *string
	Latitude        *float64
	Longitude       *float64
	Interests       *[]string
	Languages       *[]string
	Photos          *[]PhotoDTO
}

type UserUsecase struct {
	repo           repository.UserRepository
	jobRepo        repository.JobRepository
	sessionRepo    repository.SessionRepository
	asynqClient    *asynq.Client
	storageUC      *StorageUsecase
	wsDisconnecter WsDisconnecter
}

func NewUserUsecase(
	repo repository.UserRepository,
	jobRepo repository.JobRepository,
	sessionRepo repository.SessionRepository,
	asynqClient *asynq.Client,
	storageUC *StorageUsecase,
	wsDisconnecter WsDisconnecter,
) *UserUsecase {
	return &UserUsecase{
		repo:           repo,
		jobRepo:        jobRepo,
		sessionRepo:    sessionRepo,
		asynqClient:    asynqClient,
		storageUC:      storageUC,
		wsDisconnecter: wsDisconnecter,
	}
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

	if data.RelationshipType != nil {
		if *data.RelationshipType == "" {
			user.RelationshipTypeID = nil
		} else {
			uid, err := uuid.Parse(*data.RelationshipType)
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
			parts := strings.Split(*data.InterestedIn, ",")
			var newGenders []entities.MasterGender
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if uid, err := uuid.Parse(p); err == nil {
					m := entities.MasterGender{}
					m.ID = uid
					newGenders = append(newGenders, m)
				}
			}
			user.InterestedGenders = newGenders
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
				m := entities.MasterInterest{}
				m.ID = uid
				newInterests = append(newInterests, m)
			}
		}
		user.Interests = newInterests
	}

	if data.Languages != nil {
		var newLanguages []entities.MasterLanguage
		for _, rawID := range *data.Languages {
			uid, err := uuid.Parse(rawID)
			if err == nil {
				m := entities.MasterLanguage{}
				m.ID = uid
				newLanguages = append(newLanguages, m)
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
			p := entities.Photo{
				UserID:    userID,
				URL:       stripPublicURL(p.URL),
				IsMain:    p.IsMain,
				SortOrder: i,
			}
			p.ID = pid
			newPhotos = append(newPhotos, p)
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
	// 1. Authenticate user - Handled via API AuthMiddleware routing to userID
	ctx := context.Background()

	// 2. Perform Soft Delete on the user record directly
	// GORM's .Delete() applied to a SoftDeleteModel automatically updates deleted_at = NOW()
	if err := u.repo.Delete(userID); err != nil {
		return err
	}

	// 3. Insert background job into DB tracking table
	jobID := uuid.New()
	basePayload := map[string]interface{}{"job_id": jobID.String(), "user_id": userID.String()}
	payloadBytes, _ := json.Marshal(basePayload)
	
	job := &entities.Job{
		BaseModel:     entities.BaseModel{ID: jobID},
		Type:          jobs.TaskUserCleanup,
		Status:        entities.JobStatusPending,
		Payload:       payloadBytes,
		ReferenceID:   &userID,
		ReferenceType: "user",
		Source:        "user_service",
	}

	if err := u.jobRepo.CreateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to create job: %v", err)
	}

	// 4. Enqueue job to Asynq with the same payload structure
	if u.asynqClient != nil {
		taskJSON, _ := json.Marshal(jobs.UserCleanupPayload{
			BaseJobPayload: background.BaseJobPayload{JobID: jobID.String()},
			UserID:         userID,
		}) // Use struct mapping directly from background module
		
		task := asynq.NewTask(jobs.TaskUserCleanup, taskJSON)
		if _, err := u.asynqClient.Enqueue(task); err != nil {
			return fmt.Errorf("failed to enqueue background job: %v", err)
		}
	}

	// 5. Invalidate Sessions and tokens
	_ = u.sessionRepo.RevokeAllUserTokens(userID)

	// 6. Disconnect WebSocket gracefully
	if u.wsDisconnecter != nil {
		u.wsDisconnecter.DisconnectUser(userID)
	}

	return nil
}

func stripPublicURL(url string) string {
	if !strings.HasPrefix(url, "http") {
		return url
	}
	// Attempt to find the storage key part starting from users/ or chat/
	if idx := strings.Index(url, "users/"); idx != -1 {
		return url[idx:]
	}
	if idx := strings.Index(url, "chat/"); idx != -1 {
		return url[idx:]
	}
	return url
}
