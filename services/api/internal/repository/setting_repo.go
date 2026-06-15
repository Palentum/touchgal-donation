package repository

import (
	"context"
	"time"

	"donation-site/services/api/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettingRepository struct{ db *gorm.DB }

func NewSettingRepository(db *gorm.DB) *SettingRepository { return &SettingRepository{db: db} }

func (r *SettingRepository) Get(ctx context.Context, key string) (model.JSONMap, error) {
	var setting model.AppSetting
	err := r.db.WithContext(ctx).First(&setting, "key = ?", key).Error
	if err != nil {
		return nil, err
	}
	return setting.Value, nil
}

func (r *SettingRepository) Upsert(ctx context.Context, key string, value model.JSONMap) error {
	now := time.Now().UTC()
	setting := model.AppSetting{Key: key, Value: value, CreatedAt: now, UpdatedAt: now}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{"value": value, "updated_at": now}),
	}).Create(&setting).Error
}

func (r *SettingRepository) CreateIfMissing(ctx context.Context, key string, value model.JSONMap) error {
	now := time.Now().UTC()
	setting := model.AppSetting{Key: key, Value: value, CreatedAt: now, UpdatedAt: now}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&setting).Error
}
