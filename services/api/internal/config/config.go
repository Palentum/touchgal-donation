package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultSessionSecret       = "change-me-at-least-32-random-bytes"
	DefaultCSRFSecret          = "change-me-at-least-32-random-bytes"
	DefaultInitialAdminPass    = "change-me-now"
	DefaultInitialAdminPath    = "/support-console-9c2e"
	DefaultDonationMessageSize = 300
	DefaultPublicDonationDays  = 30
)

type Config struct {
	AppEnv                 string
	AppPublicURL           string
	SiteTimezone           string
	APIAddr                string
	DatabaseURL            string
	FrontendOrigin         string
	SessionSecret          string
	CSRFSecret             string
	UploadDir              string
	MaxDonationMessageLen  int
	MaxPublicDonationsDays int
	InitialAdminUsername   string
	InitialAdminPassword   string
	InitialAdminBasePath   string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		AppPublicURL:           strings.TrimRight(getEnv("APP_PUBLIC_URL", "http://localhost:3000"), "/"),
		SiteTimezone:           getEnv("SITE_TIMEZONE", "Asia/Shanghai"),
		APIAddr:                getEnv("API_ADDR", ":8080"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://donation:donation@postgres:5432/donation?sslmode=disable"),
		FrontendOrigin:         strings.TrimRight(getEnv("FRONTEND_ORIGIN", "http://localhost:3000"), "/"),
		SessionSecret:          getEnv("SESSION_SECRET", DefaultSessionSecret),
		CSRFSecret:             getEnv("CSRF_SECRET", DefaultCSRFSecret),
		UploadDir:              getEnv("UPLOAD_DIR", "/data/uploads"),
		MaxDonationMessageLen:  getEnvInt("MAX_DONATION_MESSAGE_LEN", DefaultDonationMessageSize),
		MaxPublicDonationsDays: getEnvInt("MAX_PUBLIC_DONATIONS_DAYS", DefaultPublicDonationDays),
		InitialAdminUsername:   getEnv("INITIAL_ADMIN_USERNAME", "admin"),
		InitialAdminPassword:   getEnv("INITIAL_ADMIN_PASSWORD", DefaultInitialAdminPass),
		InitialAdminBasePath:   getEnv("INITIAL_ADMIN_BASE_PATH", DefaultInitialAdminPath),
	}
	return cfg, cfg.Validate()
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

func (c Config) CookieSecure() bool {
	return c.IsProduction() || strings.HasPrefix(c.FrontendOrigin, "https://")
}

func (c Config) Validate() error {
	var errs []string
	if strings.TrimSpace(c.DatabaseURL) == "" {
		errs = append(errs, "DATABASE_URL is required")
	}
	if len(c.SessionSecret) < 32 {
		errs = append(errs, "SESSION_SECRET must be at least 32 bytes")
	}
	if len(c.CSRFSecret) < 32 {
		errs = append(errs, "CSRF_SECRET must be at least 32 bytes")
	}
	if c.MaxDonationMessageLen <= 0 {
		errs = append(errs, "MAX_DONATION_MESSAGE_LEN must be positive")
	}
	if c.MaxPublicDonationsDays <= 0 || c.MaxPublicDonationsDays > 365 {
		errs = append(errs, "MAX_PUBLIC_DONATIONS_DAYS must be between 1 and 365")
	}
	if strings.TrimSpace(c.InitialAdminUsername) == "" {
		errs = append(errs, "INITIAL_ADMIN_USERNAME is required")
	}
	if c.IsProduction() {
		if c.SessionSecret == DefaultSessionSecret {
			errs = append(errs, "SESSION_SECRET must be changed in production")
		}
		if c.CSRFSecret == DefaultCSRFSecret {
			errs = append(errs, "CSRF_SECRET must be changed in production")
		}
		if c.InitialAdminPassword == DefaultInitialAdminPass || len(c.InitialAdminPassword) < 12 {
			errs = append(errs, "INITIAL_ADMIN_PASSWORD must be changed to at least 12 characters in production")
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func (c Config) String() string {
	return fmt.Sprintf("env=%s addr=%s frontend=%s upload_dir=%s", c.AppEnv, c.APIAddr, c.FrontendOrigin, c.UploadDir)
}
