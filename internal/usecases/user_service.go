package usecases

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/repository"
)

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
