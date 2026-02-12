package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"freekiosk-hub/internal/api"
	"freekiosk-hub/internal/clients"
	"freekiosk-hub/internal/config"
	"freekiosk-hub/internal/databases"
	"freekiosk-hub/internal/network"
	"freekiosk-hub/internal/repositories"
	"freekiosk-hub/internal/services"
)

func main() {
	// 1. Configuration & Logger initialization
	cfg := config.Load()

	slog.Info("🚀 Starting FreeKiosk Hub",
		"port", cfg.ServerPort,
		"db_path", cfg.DBPath,
	)

	// Global context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var httpClient *http.Client

	// 2. Network Management (Tailscale vs Standard)
	if cfg.TSAuthKey != "" {
		slog.Info("🔐 Tailscale auth key detected, connecting to tailnet...")

		tsNode, err := network.InitTailscale(cfg.TSAuthKey, "freekiosk-hub-server")
		if err != nil {
			slog.Error("❌ Failed to initialize Tailscale", "error", err)
			os.Exit(1)
		}
		defer tsNode.Close()

		slog.Info("⏳ Waiting for Tailscale network to be up...")
		if _, err := tsNode.Server.Up(ctx); err != nil {
			slog.Error("❌ Could not bring Tailscale network up", "error", err)
			os.Exit(1)
		}
		slog.Info("✅ Tailscale network is operational")
		httpClient = tsNode.Client
	} else {
		slog.Warn("⚠️ No Tailscale key found. Using standard network stack.")
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	// 3. Database connection
	db, err := databases.Open(cfg.DBPath)
	if err != nil {
		slog.Error("❌ Failed to open database", "path", cfg.DBPath, "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Repositories & Clients initialization
	tabletRepo := repositories.NewTabletRepository(db)
	reportRepo := repositories.NewReportRepository(db)
	groupRepo := repositories.NewGroupRepository(db)
	kioskClient := clients.NewKioskClient(httpClient, cfg.KioskApiKey)

	// Ensure tables exist
	if err := tabletRepo.InitTable(); err != nil {
		slog.Error("❌ Failed to initialize tablets table", "error", err)
		os.Exit(1)
	}
	if err := reportRepo.InitTable(); err != nil {
		slog.Error("❌ Failed to initialize reports table", "error", err)
		os.Exit(1)
	}
	if err := groupRepo.InitTable(); err != nil {
		slog.Error("Échec initialisation table groups", "err", err)
		os.Exit(1)
	}
	slog.Info("✅ Database schema is ready")

	// 5. Monitoring Service initialization
	monitorSvc := services.NewMonitorService(
		tabletRepo,
		reportRepo,
		kioskClient,
		cfg.MaxWorkers,
		cfg.KioskPort,
		cfg.PollInterval,
		cfg.RetentionDays,
	)

	// 6. Launch Background Monitor Service
	if cfg.PollInterval > 0 {
		go func() {
			slog.Info("📡 Starting background monitoring service", "interval", cfg.PollInterval)
			if err := monitorSvc.Start(ctx); err != nil && err != context.Canceled {
				slog.Error("❌ Monitor service exited with error", "error", err)
			}
		}()
	} else {
		slog.Warn("ℹ️ Automatic monitoring is disabled (POLL_INTERVAL <= 0)")
	}

	e := echo.New()
	e.Renderer = &api.TemplRenderer{}
	api.NewRouter(e, db.DB, tabletRepo, reportRepo, groupRepo, monitorSvc, cfg.KioskApiKey)
	go func() {
		slog.Info("🌐 Web Server starting", "port", cfg.ServerPort)
		if err := e.Start(":" + cfg.ServerPort); err != nil && err != http.ErrServerClosed {
			slog.Error("❌ Server failed", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("🌐 Hub is fully operational. Waiting for interrupt signals...")
	<-ctx.Done()

	slog.Warn("⚠️ Shutdown signal received, stopping server...")

	time.Sleep(1 * time.Second)
	slog.Info("👋 Shutdown complete. Bye!")
}
