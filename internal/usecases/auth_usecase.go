package usecases

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/pkg/auth"
)

type AuthUsecase struct {
	repo repository.UserRepository
}

func NewAuthUsecase(repo repository.UserRepository) *AuthUsecase {
	return &AuthUsecase{repo: repo}
}

type GoogleLoginDTO struct {
	Email          string
	GoogleID       string
	FullName       string
	ProfilePicture string
}

func (u *AuthUsecase) LoginWithGoogle(dto GoogleLoginDTO) (string, error) {
	// 1. Check if this Google account is already linked
	user, err := u.repo.GetByProvider("google", dto.GoogleID)
	if err == nil {
		return auth.GenerateToken(user.ID)
	}

	// 2. Check if email exists (Account Linking)
	user, err = u.repo.GetByEmail(dto.Email)
	if err == nil {
		// Link existing email account to Google
		err = u.repo.LinkProvider(user.ID, "google", dto.GoogleID)
		if err != nil {
			return "", err
		}
		return auth.GenerateToken(user.ID)
	}

	// 3. Register New User via Google
	newUser := &entities.User{
		ID:        uuid.New(),
		Email:     &dto.Email,
		Status:    entities.UserStatusOnboarding,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Profile: &entities.Profile{
			FullName:    dto.FullName,
			DateOfBirth: time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Photos: []entities.Photo{
			{
				ID:        uuid.New(),
				URL:       dto.ProfilePicture,
				IsMain:    true,
				CreatedAt: time.Now(),
			},
		},
		AuthProviders: []entities.AuthProvider{
			{
				ID:             uuid.New(),
				Provider:       "google",
				ProviderUserID: dto.GoogleID,
				CreatedAt:      time.Now(),
			},
		},
	}

	err = u.repo.CreateWithProfile(newUser)
	if err != nil {
		return "", err
	}

	return auth.GenerateToken(newUser.ID)
}

func (u *AuthUsecase) CheckEmail(email string) (bool, error) {
	_, err := u.repo.GetByEmail(email)
	if err != nil {
		return false, nil // Email not found
	}
	return true, nil // Email exists
}

func (u *AuthUsecase) RegisterEmail(email, password, fullName string, dateOfBirth time.Time) (string, error) {
	exists, _ := u.CheckEmail(email)
	if exists {
		return "", errors.New("email already registered")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := &entities.User{
		ID:           uuid.New(),
		Email:        &email,
		PasswordHash: &hashedPassword,
		Status:       entities.UserStatusOnboarding,
		CreatedAt:    time.Now(),
		Profile: &entities.Profile{
			FullName:    fullName,
			DateOfBirth: dateOfBirth,
			CreatedAt:   time.Now(),
		},
	}

	err = u.repo.CreateWithProfile(user)
	if err != nil {
		return "", err
	}

	return auth.GenerateToken(user.ID)
}

func (u *AuthUsecase) LoginEmail(email, password string) (string, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("email or password incorrect")
	}

	if user.PasswordHash == nil || !auth.CheckPasswordHash(password, *user.PasswordHash) {
		return "", errors.New("email or password incorrect")
	}

	return auth.GenerateToken(user.ID)
}
