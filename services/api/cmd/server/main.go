package main

import (
	"context"
	"log"
	"os"

	"donation-site/services/api/internal/config"
	"donation-site/services/api/internal/db"
	httpapi "donation-site/services/api/internal/http"
	"donation-site/services/api/internal/service"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o750); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}
	gormDB, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := db.Migrate(ctx, gormDB); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	adminSvc := service.NewAdminService(gormDB, cfg)
	if err := adminSvc.Seed(ctx); err != nil {
		log.Fatalf("seed database: %v", err)
	}
	authSvc := service.NewAuthService(gormDB, cfg)
	donationSvc := service.NewDonationService(gormDB, cfg)
	exportSvc := service.NewExportService(gormDB)
	overviewSvc := service.NewOverviewService(gormDB)
	routeSvc := service.NewRouteService(gormDB, cfg)
	app := httpapi.NewRouter(httpapi.Dependencies{Config: cfg, DB: gormDB, Donation: donationSvc, Auth: authSvc, Admin: adminSvc, Export: exportSvc, Overview: overviewSvc, Route: routeSvc})
	log.Printf("starting donation api: %s", cfg.String())
	if err := app.Listen(cfg.APIAddr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
