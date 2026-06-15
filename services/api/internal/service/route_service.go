package service

import (
	"context"
	"regexp"
	"strings"

	"donation-site/services/api/internal/config"
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/repository"

	"gorm.io/gorm"
)

var adminPathPattern = regexp.MustCompile(`^/[A-Za-z0-9_/-]+$`)

type RouteService struct {
	cfg      config.Config
	settings *repository.SettingRepository
}

type RouteResolution struct {
	Kind     string `json:"kind"`
	BasePath string `json:"base_path"`
	SubPath  string `json:"sub_path"`
}

func NewRouteService(db *gorm.DB, cfg config.Config) *RouteService {
	return &RouteService{cfg: cfg, settings: repository.NewSettingRepository(db)}
}

func ValidateAdminBasePath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return Validation("后台路径必须以 / 开头", nil)
	}
	if len(path) < 6 || len(path) > 80 {
		return Validation("后台路径长度必须在 6 到 80 之间", nil)
	}
	if !adminPathPattern.MatchString(path) {
		return Validation("后台路径只能包含字母、数字、-、_ 和 /", nil)
	}
	blocked := []string{"/", "/api", "/thanks", "/assets", "/_nuxt"}
	for _, item := range blocked {
		if path == item || strings.HasPrefix(path, item+"/") {
			return Validation("后台路径与系统路径冲突", nil)
		}
	}
	return nil
}

func ResolveAdminPath(path, basePath string) (RouteResolution, bool) {
	if path == basePath {
		return RouteResolution{Kind: "admin", BasePath: basePath, SubPath: "/"}, true
	}
	if strings.HasPrefix(path, basePath+"/") {
		return RouteResolution{Kind: "admin", BasePath: basePath, SubPath: strings.TrimPrefix(path, basePath)}, true
	}
	return RouteResolution{}, false
}

func (s *RouteService) Resolve(ctx context.Context, path string) (RouteResolution, error) {
	admin, err := s.settings.Get(ctx, "admin")
	if err != nil {
		return RouteResolution{}, err
	}
	base, _ := admin["base_path"].(string)
	if base == "" {
		base = s.cfg.InitialAdminBasePath
	}
	if result, ok := ResolveAdminPath(path, base); ok {
		return result, nil
	}
	return RouteResolution{}, NotFound("路径不存在")
}

func AdminSettingValue(basePath string) model.JSONMap {
	return model.JSONMap{"base_path": basePath}
}
