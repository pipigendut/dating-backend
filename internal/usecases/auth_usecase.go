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
	repo        repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthUsecase(repo repository.UserRepository, sessionRepo repository.SessionRepository) *AuthUsecase {
	return &AuthUsecase{repo: repo, sessionRepo: sessionRepo}
}

type PhotoDTO struct {
	URL    string
	IsMain bool
}

type DeviceDTO struct {
	DeviceID    string
	DeviceName  string
	DeviceModel string
	OSVersion   string
	AppVersion  string
	FCMToken    string
	LastIP      string
}

type GoogleLoginDTO struct {
	Email           string
	GoogleID        string
	FullName        string
	ProfilePicture  string
	DateOfBirth     *string
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
	Device          DeviceDTO
}

type RegisterEmailDTO struct {
	Email           string
	Password        string
	FullName        string
	DateOfBirth     string
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
	Device          DeviceDTO
}

type LoginEmailDTO struct {
	Email    string
	Password string
	Device   DeviceDTO
}

func (u *AuthUsecase) generateTokensAndDevice(userID uuid.UUID, dto DeviceDTO) (string, string, error) {
	accessToken, err := auth.GenerateToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	hashedToken := auth.HashToken(refreshTokenStr)

	// Save or update Device
	deviceID := dto.DeviceID
	if deviceID == "" {
		deviceID = uuid.NewString() // Fallback if client doesn't provide one
	}

	device, err := u.sessionRepo.GetDeviceByDeviceIDAndUserID(deviceID, userID)
	if err != nil {
		device = &entities.Device{
			ID:          uuid.New(),
			UserID:      userID,
			DeviceID:    deviceID,
			DeviceName:  dto.DeviceName,
			DeviceModel: dto.DeviceModel,
			OSVersion:   dto.OSVersion,
			AppVersion:  dto.AppVersion,
			LastIP:      dto.LastIP,
			LastLogin:   time.Now(),
			IsActive:    true,
		}
		if dto.FCMToken != "" {
			device.FCMToken = &dto.FCMToken
		}
		_ = u.sessionRepo.CreateDevice(device)
	} else {
		device.LastLogin = time.Now()
		device.IsActive = true
		device.DeviceName = dto.DeviceName
		device.OSVersion = dto.OSVersion
		if dto.FCMToken != "" {
			device.FCMToken = &dto.FCMToken
		}
		_ = u.sessionRepo.UpdateDevice(device)
	}

	// Save RefreshToken
	rf := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		DeviceID:  device.ID,
		TokenHash: hashedToken,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}
	err = u.sessionRepo.CreateRefreshToken(rf)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenStr, nil
}

func (u *AuthUsecase) LoginWithGoogle(dto GoogleLoginDTO) (string, string, *entities.User, error) {
	// 1. Check if this Google account is already linked
	user, err := u.repo.GetByProvider("google", dto.GoogleID)
	if err == nil {
		token, refresh, err := u.generateTokensAndDevice(user.ID, dto.Device)
		return token, refresh, user, err
	}

	// 2. Check if email exists (Account Linking)
	user, err = u.repo.GetByEmail(dto.Email)
	if err == nil {
		// Link existing email account to Google
		err = u.repo.LinkProvider(user.ID, "google", dto.GoogleID)
		if err != nil {
			return "", "", nil, err
		}
		token, refresh, err := u.generateTokensAndDevice(user.ID, dto.Device)
		return token, refresh, user, err
	}

	// 3. Register New User via Google
	userID := uuid.New()

	dob := time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)
	if dto.DateOfBirth != nil {
		if t, err := time.Parse("2006-01-02", *dto.DateOfBirth); err == nil {
			dob = t
		}
	}

	status := entities.UserStatusOnboarding
	// If it's from the last step, it should be active
	if dto.Languages != nil && len(*dto.Languages) > 0 {
		status = entities.UserStatusActive
	}

	newUser := &entities.User{
		ID:        userID,
		Email:     &dto.Email,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Profile: &entities.Profile{
			UserID:          userID,
			FullName:        dto.FullName,
			DateOfBirth:     dob,
			Gender:          getString(dto.Gender),
			HeightCM:        getInt(dto.HeightCM),
			Bio:             getString(dto.Bio),
			InterestedIn:    getString(dto.InterestedIn),
			LookingFor:      getString(dto.LookingFor),
			LocationCity:    getString(dto.LocationCity),
			LocationCountry: getString(dto.LocationCountry),
			Latitude:        dto.Latitude,
			Longitude:       dto.Longitude,
			Interests:       joinStrings(dto.Interests),
			Languages:       joinStrings(dto.Languages),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	if dto.Photos != nil && len(*dto.Photos) > 0 {
		for i, p := range *dto.Photos {
			newUser.Photos = append(newUser.Photos, entities.Photo{
				ID:        uuid.New(),
				UserID:    userID,
				URL:       p.URL,
				IsMain:    p.IsMain,
				SortOrder: i,
				CreatedAt: time.Now(),
			})
		}
	} else if dto.ProfilePicture != "" {
		newUser.Photos = []entities.Photo{
			{
				ID:        uuid.New(),
				UserID:    userID,
				URL:       dto.ProfilePicture,
				IsMain:    true,
				CreatedAt: time.Now(),
			},
		}
	}

	newUser.AuthProviders = []entities.AuthProvider{
		{
			ID:             uuid.New(),
			UserID:         userID,
			Provider:       "google",
			ProviderUserID: dto.GoogleID,
			CreatedAt:      time.Now(),
		},
	}

	err = u.repo.CreateWithProfile(newUser)
	if err != nil {
		return "", "", nil, err
	}

	token, refresh, err := u.generateTokensAndDevice(newUser.ID, dto.Device)
	return token, refresh, newUser, err
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func joinStrings(s *[]string) string {
	if s == nil || len(*s) == 0 {
		return ""
	}
	res := ""
	for i, v := range *s {
		if i > 0 {
			res += ","
		}
		res += v
	}
	return res
}

func (u *AuthUsecase) CheckEmail(email string) (bool, error) {
	_, err := u.repo.GetByEmail(email)
	if err != nil {
		return false, nil // Email not found
	}
	return true, nil // Email exists
}

func (u *AuthUsecase) RegisterEmail(dto RegisterEmailDTO) (string, string, *entities.User, error) {
	exists, _ := u.CheckEmail(dto.Email)
	if exists {
		return "", "", nil, errors.New("email already registered")
	}

	hashedPassword, err := auth.HashPassword(dto.Password)
	if err != nil {
		return "", "", nil, err
	}

	userID := uuid.New()
	dob, _ := time.Parse("2006-01-02", dto.DateOfBirth)

	status := entities.UserStatusOnboarding
	if dto.Languages != nil && len(*dto.Languages) > 0 {
		status = entities.UserStatusActive
	}

	user := &entities.User{
		ID:           userID,
		Email:        &dto.Email,
		PasswordHash: &hashedPassword,
		Status:       status,
		CreatedAt:    time.Now(),
		Profile: &entities.Profile{
			UserID:          userID,
			FullName:        dto.FullName,
			DateOfBirth:     dob,
			Gender:          getString(dto.Gender),
			HeightCM:        getInt(dto.HeightCM),
			Bio:             getString(dto.Bio),
			InterestedIn:    getString(dto.InterestedIn),
			LookingFor:      getString(dto.LookingFor),
			LocationCity:    getString(dto.LocationCity),
			LocationCountry: getString(dto.LocationCountry),
			Latitude:        dto.Latitude,
			Longitude:       dto.Longitude,
			Interests:       joinStrings(dto.Interests),
			Languages:       joinStrings(dto.Languages),
			CreatedAt:       time.Now(),
		},
	}

	if dto.Photos != nil && len(*dto.Photos) > 0 {
		for i, p := range *dto.Photos {
			user.Photos = append(user.Photos, entities.Photo{
				ID:        uuid.New(),
				UserID:    userID,
				URL:       p.URL,
				IsMain:    p.IsMain,
				SortOrder: i,
				CreatedAt: time.Now(),
			})
		}
	}

	err = u.repo.CreateWithProfile(user)
	if err != nil {
		return "", "", nil, err
	}

	token, refresh, err := u.generateTokensAndDevice(user.ID, dto.Device)
	return token, refresh, user, err
}

func (u *AuthUsecase) LoginEmail(dto LoginEmailDTO) (string, string, *entities.User, error) {
	user, err := u.repo.GetByEmail(dto.Email)
	if err != nil {
		return "", "", nil, errors.New("email or password incorrect")
	}

	if user.PasswordHash == nil || !auth.CheckPasswordHash(dto.Password, *user.PasswordHash) {
		return "", "", nil, errors.New("email or password incorrect")
	}

	token, refresh, err := u.generateTokensAndDevice(user.ID, dto.Device)
	return token, refresh, user, err
}

func (u *AuthUsecase) RefreshToken(refreshTokenStr, deviceID string) (string, string, error) {
	hashedToken := auth.HashToken(refreshTokenStr)
	tokenRec, err := u.sessionRepo.GetRefreshTokenByHash(hashedToken)
	if err != nil {
		return "", "", errors.New("invalid or expired refresh token")
	}

	if tokenRec.RevokedAt != nil || tokenRec.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token is expired or revoked")
	}

	device, err := u.sessionRepo.GetDeviceByDeviceIDAndUserID(deviceID, tokenRec.UserID)
	if err != nil || !device.IsActive {
		return "", "", errors.New("invalid or inactive device")
	}

	_ = u.sessionRepo.RevokeRefreshToken(tokenRec.ID)

	return u.generateTokensAndDevice(tokenRec.UserID, DeviceDTO{
		DeviceID:    deviceID,
		DeviceName:  device.DeviceName,
		DeviceModel: device.DeviceModel,
		OSVersion:   device.OSVersion,
		AppVersion:  device.AppVersion,
		LastIP:      device.LastIP,
	})
}

func (u *AuthUsecase) Logout(deviceID string, userID uuid.UUID) error {
	_ = u.sessionRepo.RevokeAllUserTokens(userID)
	return u.sessionRepo.DeactivateDevice(deviceID, userID)
}
