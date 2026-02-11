package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"freekiosk-hub/internal/clients"
	"freekiosk-hub/internal/repositories"
)

type MonitorService interface {
	Start(ctx context.Context) error
	ScanAll()
}

type monitorServiceImpl struct {
	tabletRepo    repositories.TabletRepository
	reportRepo    repositories.ReportRepository
	kioskClient   clients.KioskClient
	maxWorkers    int
	kioskPort     string
	pollInterval  time.Duration
	retentionDays int
}

func NewMonitorService(
	tr repositories.TabletRepository,
	rr repositories.ReportRepository,
	kc clients.KioskClient,
	maxWorkers int,
	kioskPort string,
	pollInterval time.Duration,
	retentionDays int,
) MonitorService {
	return &monitorServiceImpl{
		tabletRepo:    tr,
		reportRepo:    rr,
		kioskClient:   kc,
		maxWorkers:    maxWorkers,
		kioskPort:     kioskPort,
		pollInterval:  pollInterval,
		retentionDays: retentionDays,
	}
}

func (s *monitorServiceImpl) ScanAll() {
	tablets, err := s.tabletRepo.GetAll()
	if err != nil {
		slog.Error("Failed to fetch tablets from DB", "error", err)
		return
	}

	if len(tablets) == 0 {
		slog.Info("No tablets found in database for scanning")
		return
	}

	slog.Info("Starting global scan", "count", len(tablets), "workers", s.maxWorkers)

	jobs := make(chan repositories.Tablet, len(tablets))
	var wg sync.WaitGroup

	for w := 1; w <= s.maxWorkers; w++ {
		wg.Add(1)
		go s.worker(&wg, jobs)
	}

	for _, t := range tablets {
		jobs <- t
	}
	close(jobs)

	wg.Wait()
	slog.Info("Global scan completed")

	if s.retentionDays > 0 {
		slog.Info("Starting reports cleanup", "retention_days", s.retentionDays)
		if err := s.reportRepo.Cleanup(s.retentionDays); err != nil {
			slog.Error("Failed to cleanup old reports", "error", err)
		} else {
			slog.Info("Reports cleanup finished")
		}
	}
}

func (s *monitorServiceImpl) worker(wg *sync.WaitGroup, jobs <-chan repositories.Tablet) {
	defer wg.Done()

	for t := range jobs {
		host := fmt.Sprintf("%s:%s", t.IP, s.kioskPort)

		report, err := s.kioskClient.FetchStatus(host)

		report.TabletID = t.ID

		if err == nil && report.Success {
			t.Online = true
			t.LastSeen = time.Now()
			t.Version = report.DeviceVersion
		} else {
			t.Online = false
			slog.Info("Tablet offline or returned error", "id", t.ID, "ip", t.IP, "error", err)
		}

		if err := s.tabletRepo.Save(&t); err != nil {
			slog.Error("Failed to update tablet status", "id", t.ID, "error", err)
		}

		if err := s.reportRepo.Add(report); err != nil {
			slog.Error("Failed to save report", "id", t.ID, "error", err)
		}
	}
}

func (s *monitorServiceImpl) Start(ctx context.Context) error {
	slog.Info("Starting monitor service", "interval", s.pollInterval)

	// On lance un premier scan immédiatement au démarrage
	s.ScanAll()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Debug("Ticker ticked, starting scheduled scan")
			s.ScanAll()
		case <-ctx.Done():
			slog.Info("Monitor service shutting down")
			return ctx.Err()
		}
	}
}
