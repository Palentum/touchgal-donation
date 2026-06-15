package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"donation-site/services/api/internal/config"
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/repository"
	"donation-site/services/api/internal/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const AdminSessionCookie = "admin_session"

type AuthService struct {
	cfg    config.Config
	admins *repository.AdminRepository
}

type LoginResult struct {
	Admin     model.Admin
	Token     string
	CSRFToken string
	ExpiresAt time.Time
}

type AuthContext struct {
	Admin   model.Admin
	Session model.AdminSession
}

func NewAuthService(db *gorm.DB, cfg config.Config) *AuthService {
	return &AuthService{cfg: cfg, admins: repository.NewAdminRepository(db)}
}

func (s *AuthService) Login(ctx context.Context, username, password, ip, userAgent string) (*LoginResult, error) {
	admin, err := s.admins.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil || admin.Status != model.AdminStatusActive || !util.VerifyPassword(password, admin.PasswordHash) {
		return nil, Unauthorized("用户名或密码错误")
	}
	token, err := util.RandomToken(32)
	if err != nil {
		return nil, err
	}
	csrfToken, err := util.RandomToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)
	session := &model.AdminSession{
		AdminID:       admin.ID,
		TokenHash:     util.HMACSHA256(s.cfg.SessionSecret, token),
		CSRFTokenHash: util.HMACSHA256(s.cfg.CSRFSecret, csrfToken),
		IPHash:        util.HashFingerprint(s.cfg.SessionSecret, ip),
		UserAgentHash: util.HashFingerprint(s.cfg.SessionSecret, userAgent),
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
	}
	if err := s.admins.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	_ = s.admins.UpdateLastLogin(ctx, admin.ID, now)
	admin.LastLoginAt = &now
	return &LoginResult{Admin: *admin, Token: token, CSRFToken: csrfToken, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (*AuthContext, error) {
	if token == "" {
		return nil, Unauthorized("请先登录")
	}
	session, err := s.admins.FindSessionByTokenHash(ctx, util.HMACSHA256(s.cfg.SessionSecret, token), time.Now().UTC())
	if err != nil {
		return nil, Unauthorized("请先登录")
	}
	if session.Admin.Status != model.AdminStatusActive {
		return nil, Unauthorized("请先登录")
	}
	return &AuthContext{Admin: session.Admin, Session: *session}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.admins.RevokeSession(ctx, util.HMACSHA256(s.cfg.SessionSecret, token), time.Now().UTC())
}

func (s *AuthService) VerifyCSRF(session model.AdminSession, token string) bool {
	if token == "" {
		return false
	}
	return util.ConstantEqualHex(session.CSRFTokenHash, util.HMACSHA256(s.cfg.CSRFSecret, token))
}

func (s *AuthService) ChangePassword(ctx context.Context, adminID uuid.UUID, oldPassword, newPassword string) error {
	admin, err := s.admins.FindByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound("管理员不存在")
		}
		return err
	}
	if !util.VerifyPassword(oldPassword, admin.PasswordHash) {
		return Unauthorized("旧密码错误")
	}
	if len([]rune(newPassword)) < 12 {
		return Validation("新密码至少 12 位", nil)
	}
	if strings.EqualFold(newPassword, admin.Username) {
		return Validation("新密码不能与用户名相同", nil)
	}
	hash, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.admins.UpdatePassword(ctx, adminID, hash)
}
