package impl

import (
	"context"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type entityRepository struct {
	db *gorm.DB
}

func NewEntityRepository(db *gorm.DB) repository.EntityRepository {
	return &entityRepository{db: db}
}

func (r *entityRepository) Create(ctx context.Context, entity *entities.Entity) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *entityRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Entity, error) {
	var ent entities.Entity
	query := r.db.WithContext(ctx)
	query = ApplyFullUserPreload(query, "User")
	query = ApplyFullUserPreload(query, "Group.Members.User")
	
	err := query.First(&ent, "id = ?", id).Error
	return &ent, err
}

func (r *entityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Entity{}, "id = ?", id).Error
}
