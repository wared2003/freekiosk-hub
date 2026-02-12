package services

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"freekiosk-hub/internal/clients"
	"freekiosk-hub/internal/repositories"
)

var (
	ErrTabletNotFound   = errors.New("tablet_not_found")
	ErrGroupNotFound    = errors.New("group_not_found")
	ErrInvalidTarget    = errors.New("invalid_target_specification")
	ErrKioskUnreachable = errors.New("kiosk_unreachable")
)

type Target struct {
	TabletID int64
	GroupID  int64
	IPs      []string
}

type TabletResult struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Success  bool   `json:"success"`
	Executed bool   `json:"executed"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

type ActionReport struct {
	Command   string         `json:"command"`
	Timestamp int64          `json:"timestamp"`
	Summary   string         `json:"summary"`
	Results   []TabletResult `json:"results"`
}

type StatusResult struct {
	ID      int64                      `json:"id"`
	Name    string                     `json:"name"`
	Success bool                       `json:"success"`
	Data    *repositories.TabletReport `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

type StatusReport struct {
	Timestamp int64          `json:"timestamp"`
	Results   []StatusResult `json:"results"`
}

type KioskService interface {
	// Affichage & UI
	SetBrightness(t Target, val int) (*ActionReport, error)
	SetVolume(t Target, vol int) (*ActionReport, error)
	ShowToast(t Target, text string) (*ActionReport, error)

	// Système & Écran
	SetScreen(t Target, on bool) (*ActionReport, error)
	SetScreensaver(t Target, active bool) (*ActionReport, error)
	Wake(t Target) (*ActionReport, error)
	Reboot(t Target) (*ActionReport, error)

	// Navigation & Webview
	Navigate(t Target, url string) (*ActionReport, error)
	NavigateAlias(t Target, url string) (*ActionReport, error)
	Reload(t Target) (*ActionReport, error)
	ClearCache(t Target) (*ActionReport, error)
	ExecuteJS(t Target, code string) (*ActionReport, error)
	SetRotation(t Target, start bool) (*ActionReport, error)

	// Audio & TTS
	Speak(t Target, text string) (*ActionReport, error)
	PlayAudio(t Target, url string, loop bool, volume int) (*ActionReport, error)
	StopAudio(t Target) (*ActionReport, error)
	Beep(t Target) (*ActionReport, error)

	// Apps & Remote
	LaunchApp(t Target, packageName string) (*ActionReport, error)
	SendRemoteCommand(t Target, action string) (*ActionReport, error)

	// Média Spécifique (Photo)
	GetPhoto(tabletID int64, camera string, quality int) ([]byte, error)
	FetchStatus(t Target) (*StatusReport, error)
}

type kioskServiceImpl struct {
	tabRepo   repositories.TabletRepository
	groupRepo repositories.GroupRepository
	client    clients.KioskClient
	kPort     string
}

func NewKioskService(r repositories.TabletRepository, gr repositories.GroupRepository, c clients.KioskClient, kp string) KioskService {
	return &kioskServiceImpl{tabRepo: r, groupRepo: gr, client: c, kPort: kp}
}

func (s *kioskServiceImpl) getAddr(ip string) string {
	return fmt.Sprintf("%s:%s", ip, s.kPort)
}

// resolveTablets transforme une Target en liste d'objets tablettes réels
func (s *kioskServiceImpl) resolveTablets(t Target) ([]repositories.Tablet, error) {
	if len(t.IPs) > 0 {
		tabs := make([]repositories.Tablet, len(t.IPs))
		for i, ip := range t.IPs {
			tabs[i] = repositories.Tablet{IP: ip, Name: "Target IP"}
		}
		return tabs, nil
	}
	if t.TabletID > 0 {
		tab, err := s.tabRepo.GetByID(t.TabletID)
		if err != nil {
			return nil, ErrTabletNotFound
		}
		return []repositories.Tablet{*tab}, nil
	}
	if t.GroupID > 0 {
		tablets, err := s.groupRepo.GetTabletsByGroup(t.GroupID)
		if err != nil || len(tablets) == 0 {
			return nil, ErrGroupNotFound
		}
		return tablets, nil
	}
	return nil, ErrInvalidTarget
}

// executeAndWait est le moteur centralisé de parallélisme
func (s *kioskServiceImpl) executeAndWait(t Target, cmdName string, action func(ip string) error) (*ActionReport, error) {
	tablets, err := s.resolveTablets(t)
	if err != nil {
		return nil, err
	}

	report := &ActionReport{
		Command:   cmdName,
		Timestamp: time.Now().Unix(),
		Results:   make([]TabletResult, len(tablets)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, tab := range tablets {
		wg.Add(1)
		go func(index int, tablet repositories.Tablet) {
			defer wg.Done()
			start := time.Now()

			fullAddr := s.getAddr(tablet.IP)
			err := action(fullAddr)
			duration := time.Since(start).Round(time.Millisecond).String()

			res := TabletResult{
				ID:       tablet.ID,
				Name:     tablet.Name,
				IP:       tablet.IP,
				Duration: duration,
			}

			if err != nil {
				res.Success = false
				res.Executed = false
				res.Error = err.Error()
				slog.Warn("Action failed", "tablet", tablet.Name, "cmd", cmdName, "err", err)
			} else {
				res.Success = true
				res.Executed = true
			}

			mu.Lock()
			report.Results[index] = res
			mu.Unlock()
		}(i, tab)
	}

	wg.Wait()

	successCount := 0
	for _, r := range report.Results {
		if r.Executed {
			successCount++
		}
	}
	report.Summary = fmt.Sprintf("%d/%d tablettes ont exécuté la commande avec succès", successCount, len(tablets))

	return report, nil
}

func (s *kioskServiceImpl) FetchStatus(t Target) (*StatusReport, error) {
	tablets, err := s.resolveTablets(t)
	if err != nil {
		return nil, err
	}

	report := &StatusReport{
		Timestamp: time.Now().Unix(),
		Results:   make([]StatusResult, len(tablets)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, tab := range tablets {
		wg.Add(1)
		go func(index int, tablet repositories.Tablet) {
			defer wg.Done()

			// Appel du client FetchStatus (celui que tu as écrit)
			fullAddr := s.getAddr(tablet.IP)
			data, err := s.client.FetchStatus(fullAddr)

			res := StatusResult{
				ID:   tablet.ID,
				Name: tablet.Name,
			}

			if err != nil {
				res.Success = false
				res.Error = err.Error()
			} else {
				res.Success = data.Success
				res.Data = data
			}

			mu.Lock()
			report.Results[index] = res
			mu.Unlock()
		}(i, tab)
	}

	wg.Wait()
	return report, nil
}

// --- IMPLÉMENTATION DES MÉTHODES ---

func (s *kioskServiceImpl) SetBrightness(t Target, val int) (*ActionReport, error) {
	return s.executeAndWait(t, "setBrightness", func(ip string) error { return s.client.SetBrightness(ip, val) })
}

func (s *kioskServiceImpl) SetVolume(t Target, vol int) (*ActionReport, error) {
	return s.executeAndWait(t, "setVolume", func(ip string) error { return s.client.SetVolume(ip, vol) })
}

func (s *kioskServiceImpl) ShowToast(t Target, text string) (*ActionReport, error) {
	return s.executeAndWait(t, "showToast", func(ip string) error { return s.client.ShowToast(ip, text) })
}

func (s *kioskServiceImpl) SetScreen(t Target, on bool) (*ActionReport, error) {
	return s.executeAndWait(t, "setScreen", func(ip string) error { return s.client.SetScreen(ip, on) })
}

func (s *kioskServiceImpl) SetScreensaver(t Target, active bool) (*ActionReport, error) {
	return s.executeAndWait(t, "setScreensaver", func(ip string) error { return s.client.SetScreensaver(ip, active) })
}

func (s *kioskServiceImpl) Wake(t Target) (*ActionReport, error) {
	return s.executeAndWait(t, "wake", s.client.Wake)
}

func (s *kioskServiceImpl) Reboot(t Target) (*ActionReport, error) {
	return s.executeAndWait(t, "reboot", s.client.Reboot)
}

func (s *kioskServiceImpl) Navigate(t Target, url string) (*ActionReport, error) {
	return s.executeAndWait(t, "navigate", func(ip string) error { return s.client.Navigate(ip, url) })
}

func (s *kioskServiceImpl) NavigateAlias(t Target, url string) (*ActionReport, error) {
	return s.executeAndWait(t, "navigateAlias", func(ip string) error { return s.client.NavigateAlias(ip, url) })
}

func (s *kioskServiceImpl) Reload(t Target) (*ActionReport, error) {
	return s.executeAndWait(t, "reload", s.client.Reload)
}

func (s *kioskServiceImpl) ClearCache(t Target) (*ActionReport, error) {
	return s.executeAndWait(t, "clearCache", s.client.ClearCache)
}

func (s *kioskServiceImpl) ExecuteJS(t Target, code string) (*ActionReport, error) {
	return s.executeAndWait(t, "executeJS", func(ip string) error { return s.client.ExecuteJS(ip, code) })
}

func (s *kioskServiceImpl) SetRotation(t Target, start bool) (*ActionReport, error) {
	return s.executeAndWait(t, "setRotation", func(ip string) error { return s.client.SetRotation(ip, start) })
}

func (s *kioskServiceImpl) Speak(t Target, text string) (*ActionReport, error) {
	return s.executeAndWait(t, "speak", func(ip string) error { return s.client.Speak(ip, text) })
}

func (s *kioskServiceImpl) PlayAudio(t Target, url string, loop bool, volume int) (*ActionReport, error) {
	return s.executeAndWait(t, "playAudio", func(ip string) error { return s.client.PlayAudio(ip, url, loop, volume) })
}

func (s *kioskServiceImpl) StopAudio(t Target) (*ActionReport, error) {
	return s.executeAndWait(t, "stopAudio", s.client.StopAudio)
}

func (s *kioskServiceImpl) Beep(t Target) (*ActionReport, error) {
	return s.executeAndWait(t, "beep", s.client.Beep)
}

func (s *kioskServiceImpl) LaunchApp(t Target, packageName string) (*ActionReport, error) {
	return s.executeAndWait(t, "launchApp", func(ip string) error { return s.client.LaunchApp(ip, packageName) })
}

func (s *kioskServiceImpl) SendRemoteCommand(t Target, action string) (*ActionReport, error) {
	return s.executeAndWait(t, "remoteCommand", func(ip string) error { return s.client.SendRemoteCommand(ip, action) })
}

func (s *kioskServiceImpl) GetPhoto(tabletID int64, camera string, quality int) ([]byte, error) {
	tab, err := s.tabRepo.GetByID(tabletID)
	if err != nil {
		return nil, ErrTabletNotFound
	}
	return s.client.TakePhoto(tab.IP, camera, quality)
}
