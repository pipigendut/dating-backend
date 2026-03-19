package impl

import (
	"context"

	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type configRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) repository.ConfigRepository {
	return &configRepository{db: db}
}

func (r *configRepository) GetAll(ctx context.Context) ([]entities.AppConfig, error) {
	var configs []entities.AppConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}

func (r *configRepository) Get(ctx context.Context, key string) (*entities.AppConfig, error) {
	var config entities.AppConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *configRepository) Set(ctx context.Context, key, value string) error {
	config := entities.AppConfig{Key: key, Value: value}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&config).Error
}
func (r *configRepository) DeleteAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entities.AppConfig{}).Error
}
