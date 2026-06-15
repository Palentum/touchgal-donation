package repository

import (
	"context"
	"time"

	"donation-site/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminRepository struct{ db *gorm.DB }

func NewAdminRepository(db *gorm.DB) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Admin{}).Count(&count).Error
	return count, err
}

func (r *AdminRepository) Create(ctx context.Context, admin *model.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).First(&admin, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Admin{}).Where("id = ?", id).Updates(map[string]any{"last_login_at": at, "updated_at": at}).Error
}

func (r *AdminRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&model.Admin{}).Where("id = ?", id).Updates(map[string]any{"password_hash": hash, "must_change_password": false, "updated_at": now}).Error
}

func (r *AdminRepository) CreateSession(ctx context.Context, session *model.AdminSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *AdminRepository) FindSessionByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*model.AdminSession, error) {
	var session model.AdminSession
	err := r.db.WithContext(ctx).Preload("Admin").Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, now).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AdminRepository) RevokeSession(ctx context.Context, tokenHash string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&model.AdminSession{}).Where("token_hash = ? AND revoked_at IS NULL", tokenHash).Update("revoked_at", at).Error
}
