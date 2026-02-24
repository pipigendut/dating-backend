package usecases

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type MasterUsecase struct {
	repo repository.MasterRepository
}

func NewMasterUsecase(repo repository.MasterRepository) *MasterUsecase {
	return &MasterUsecase{repo: repo}
}

type GenderDTO struct {
	ID   uuid.UUID
	Name string
	Icon string
}

type RelationshipTypeDTO struct {
	ID   uuid.UUID
	Name string
	Icon string
}

type InterestDTO struct {
	ID   uuid.UUID
	Name string
	Icon string
}

type LanguageDTO struct {
	ID   uuid.UUID
	Name string
	Icon string
}

func (u *MasterUsecase) GetGenders() ([]GenderDTO, error) {
	data, err := u.repo.GetGenders()
	if err != nil {
		return nil, err
	}
	var dtos []GenderDTO
	for _, d := range data {
		dtos = append(dtos, GenderDTO{ID: d.ID, Name: d.Name, Icon: d.Icon})
	}
	return dtos, nil
}

func (u *MasterUsecase) GetRelationshipTypes() ([]RelationshipTypeDTO, error) {
	data, err := u.repo.GetRelationshipTypes()
	if err != nil {
		return nil, err
	}
	var dtos []RelationshipTypeDTO
	for _, d := range data {
		dtos = append(dtos, RelationshipTypeDTO{ID: d.ID, Name: d.Name, Icon: d.Icon})
	}
	return dtos, nil
}

func (u *MasterUsecase) GetInterests() ([]InterestDTO, error) {
	data, err := u.repo.GetInterests()
	if err != nil {
		return nil, err
	}
	var dtos []InterestDTO
	for _, d := range data {
		dtos = append(dtos, InterestDTO{ID: d.ID, Name: d.Name, Icon: d.Icon})
	}
	return dtos, nil
}

func (u *MasterUsecase) GetLanguages() ([]LanguageDTO, error) {
	data, err := u.repo.GetLanguages()
	if err != nil {
		return nil, err
	}
	var dtos []LanguageDTO
	for _, d := range data {
		dtos = append(dtos, LanguageDTO{ID: d.ID, Name: d.Name, Icon: d.Icon})
	}
	return dtos, nil
}
