package httpapi

import (
	"time"

	"donation-site/services/api/internal/config"
	handler "donation-site/services/api/internal/http/handler"
	apimw "donation-site/services/api/internal/http/middleware"
	"donation-site/services/api/internal/service"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type Dependencies struct {
	Config   config.Config
	DB       *gorm.DB
	Donation *service.DonationService
	Auth     *service.AuthService
	Admin    *service.AdminService
	Export   *service.ExportService
	Overview *service.OverviewService
	Route    *service.RouteService
}

func NewRouter(deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: HandleError, BodyLimit: 3 * 1024 * 1024})
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(apimw.SecurityHeaders())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{deps.Config.FrontendOrigin},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPatch, fiber.MethodDelete, fiber.MethodOptions},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept, "X-CSRF-Token"},
		AllowCredentials: true,
	}))
	app.Get("/uploads/*", static.New(deps.Config.UploadDir))

	api := app.Group("/api/v1")
	public := handler.NewPublicHandler(deps.Donation, deps.Route)
	api.Get("/public/config", public.Config)
	api.Get("/public/donations/recent", public.RecentDonations)
	api.Post("/public/donations", apimw.RateLimit(10, time.Minute), public.CreateDonation)
	api.Get("/public/donations/:orderNo/status", public.DonationStatus)
	api.Get("/public/donations/:orderNo/qr.png", public.DonationQR)
	api.Get("/public/resolve-route", apimw.RateLimit(30, time.Minute), public.ResolveRoute)

	webhook := handler.NewWebhookHandler(deps.Donation)
	api.Post("/payments/:provider/webhook", webhook.PaymentWebhook)

	authHandler := handler.NewAdminAuthHandler(deps.Auth, deps.Config, deps.Admin)
	api.Post("/admin/auth/login", apimw.RateLimit(10, 5*time.Minute), authHandler.Login)

	admin := api.Group("/admin", apimw.AuthRequired(deps.Auth), apimw.CSRFRequired(deps.Auth))
	admin.Post("/auth/logout", authHandler.Logout)
	admin.Get("/auth/me", authHandler.Me)
	admin.Post("/auth/change-password", authHandler.ChangePassword)

	adminHandler := handler.NewAdminHandler(deps.Config, deps.DB, deps.Admin, deps.Export, deps.Overview)
	admin.Get("/tiers", adminHandler.ListTiers)
	admin.Post("/tiers", adminHandler.CreateTier)
	admin.Patch("/tiers/:id", adminHandler.UpdateTier)
	admin.Delete("/tiers/:id", adminHandler.DeleteTier)

	admin.Get("/payment-methods", adminHandler.ListPaymentMethods)
	admin.Post("/payment-methods", adminHandler.CreatePaymentMethod)
	admin.Patch("/payment-methods/:id", adminHandler.UpdatePaymentMethod)
	admin.Delete("/payment-methods/:id", adminHandler.DeletePaymentMethod)
	admin.Post("/payment-methods/:id/upload-qr", adminHandler.UploadPaymentMethodQR)

	admin.Get("/donations", adminHandler.ListDonations)
	admin.Get("/donations/export", adminHandler.ExportDonations)
	admin.Get("/donations/:id", adminHandler.GetDonation)
	admin.Patch("/donations/:id/status", adminHandler.UpdateDonationStatus)

	admin.Get("/overview", adminHandler.Overview)
	admin.Get("/settings", adminHandler.GetSettings)
	admin.Patch("/settings/site", adminHandler.UpdateSiteSettings)
	admin.Patch("/settings/admin-path", adminHandler.UpdateAdminPath)
	return app
}
