package repositories

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// TabletReport represents a full status snapshot from a device
type TabletReport struct {
	ID       int64 `db:"id"`
	TabletID int64 `db:"tablet_id"` // Matches the auto-increment ID from tablets table
	Success  bool  `db:"success"`

	// Battery
	BatteryLevel    int    `db:"battery_level"`
	BatteryCharging bool   `db:"battery_charging"`
	BatteryPlugged  string `db:"battery_plugged"`

	// Screen
	ScreenOn          bool `db:"screen_on"`
	ScreenBrightness  int  `db:"screen_brightness"`
	ScreensaverActive bool `db:"screensaver_active"`

	// Audio
	AudioVolume int `db:"audio_volume"`

	// Webview
	CurrentURL       string `db:"current_url"`
	WebviewCanGoBack bool   `db:"webview_can_go_back"`
	WebviewLoading   bool   `db:"webview_loading"`

	// Device
	DeviceIP       string `db:"device_ip"`
	DeviceHostname string `db:"device_hostname"`
	DeviceVersion  string `db:"device_version"`
	IsDeviceOwner  bool   `db:"is_device_owner"`
	KioskMode      bool   `db:"kiosk_mode"`

	// WiFi
	WifiSSID           string `db:"wifi_ssid"`
	WifiSignalStrength int    `db:"wifi_signal_strength"`
	WifiSignalLevel    int    `db:"wifi_signal_level"`
	WifiConnected      bool   `db:"wifi_connected"`
	WifiLinkSpeed      int    `db:"wifi_link_speed"`
	WifiFrequency      int    `db:"wifi_frequency"`

	// Rotation
	RotationEnabled      bool `db:"rotation_enabled"`
	RotationInterval     int  `db:"rotation_interval"`
	RotationCurrentIndex int  `db:"rotation_current_index"`

	// Sensors
	LightLevel float64 `db:"light_level"`
	Proximity  float64 `db:"proximity"`
	AccelX     float64 `db:"accel_x"`
	AccelY     float64 `db:"accel_y"`
	AccelZ     float64 `db:"accel_z"`

	// Auto Brightness
	AutoBrightnessEnabled bool    `db:"auto_brightness_enabled"`
	AutoBrightnessMin     float64 `db:"auto_brightness_min"`
	AutoBrightnessMax     float64 `db:"auto_brightness_max"`
	AutoBrightnessCurrent float64 `db:"auto_brightness_current"`

	// Storage (MB / %)
	StorageTotal     int `db:"storage_total_mb"`
	StorageAvailable int `db:"storage_available_mb"`
	StorageUsed      int `db:"storage_used_mb"`
	StorageUsedPct   int `db:"storage_used_percent"`

	// Memory (MB / %)
	MemoryTotal     int  `db:"memory_total_mb"`
	MemoryAvailable int  `db:"memory_available_mb"`
	MemoryUsed      int  `db:"memory_used_mb"`
	MemoryUsedPct   int  `db:"memory_used_percent"`
	LowMemory       bool `db:"low_memory"`

	Timestamp time.Time `db:"timestamp"`
}

type ReportRepository interface {
	InitTable() error
	Add(r *TabletReport) error
	GetLatestByTablet(tabletID int64, onlySuccess bool) (*TabletReport, error)
	GetHistory(tabletID int64, limit int) ([]TabletReport, error)
	Cleanup(days int) error
}

type sqliteReportRepo struct {
	db *sqlx.DB
}

func NewReportRepository(db *sqlx.DB) ReportRepository {
	return &sqliteReportRepo{db: db}
}

func (r *sqliteReportRepo) InitTable() error {
	query := `CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tablet_id INTEGER NOT NULL,
		success BOOLEAN,
		battery_level INTEGER, battery_charging BOOLEAN, battery_plugged TEXT,
		screen_on BOOLEAN, screen_brightness INTEGER, screensaver_active BOOLEAN,
		audio_volume INTEGER,
		current_url TEXT, webview_can_go_back BOOLEAN, webview_loading BOOLEAN,
		device_ip TEXT, device_hostname TEXT, device_version TEXT, is_device_owner BOOLEAN, kiosk_mode BOOLEAN,
		wifi_ssid TEXT, wifi_signal_strength INTEGER, wifi_signal_level INTEGER, wifi_connected BOOLEAN, wifi_link_speed INTEGER, wifi_frequency INTEGER,
		rotation_enabled BOOLEAN, rotation_interval INTEGER, rotation_current_index INTEGER,
		light_level REAL, proximity REAL, accel_x REAL, accel_y REAL, accel_z REAL,
		auto_brightness_enabled BOOLEAN, auto_brightness_min REAL, auto_brightness_max REAL, auto_brightness_current REAL,
		storage_total_mb INTEGER, storage_available_mb INTEGER, storage_used_mb INTEGER, storage_used_percent INTEGER,
		memory_total_mb INTEGER, memory_available_mb INTEGER, memory_used_mb INTEGER, memory_used_percent INTEGER, low_memory BOOLEAN,
		timestamp DATETIME,
		FOREIGN KEY(tablet_id) REFERENCES tablets(id)
	);
	CREATE INDEX IF NOT EXISTS idx_reports_tablet_id ON reports(tablet_id);`

	_, err := r.db.Exec(query)
	return err
}

func (r *sqliteReportRepo) Add(report *TabletReport) error {
	if report.Timestamp.IsZero() {
		report.Timestamp = time.Now()
	}

	query := `INSERT INTO reports (
		tablet_id, success, battery_level, battery_charging, battery_plugged,
		screen_on, screen_brightness, screensaver_active, audio_volume,
		current_url, webview_can_go_back, webview_loading,
		device_ip, device_hostname, device_version, is_device_owner, kiosk_mode,
		wifi_ssid, wifi_signal_strength, wifi_signal_level, wifi_connected, wifi_link_speed, wifi_frequency,
		rotation_enabled, rotation_interval, rotation_current_index,
		light_level, proximity, accel_x, accel_y, accel_z,
		auto_brightness_enabled, auto_brightness_min, auto_brightness_max, auto_brightness_current,
		storage_total_mb, storage_available_mb, storage_used_mb, storage_used_percent,
		memory_total_mb, memory_available_mb, memory_used_mb, memory_used_percent, low_memory,
		timestamp
	) VALUES (
		:tablet_id, :success, :battery_level, :battery_charging, :battery_plugged,
		:screen_on, :screen_brightness, :screensaver_active, :audio_volume,
		:current_url, :webview_can_go_back, :webview_loading,
		:device_ip, :device_hostname, :device_version, :is_device_owner, :kiosk_mode,
		:wifi_ssid, :wifi_signal_strength, :wifi_signal_level, :wifi_connected, :wifi_link_speed, :wifi_frequency,
		:rotation_enabled, :rotation_interval, :rotation_current_index,
		:light_level, :proximity, :accel_x, :accel_y, :accel_z,
		:auto_brightness_enabled, :auto_brightness_min, :auto_brightness_max, :auto_brightness_current,
		:storage_total_mb, :storage_available_mb, :storage_used_mb, :storage_used_percent,
		:memory_total_mb, :memory_available_mb, :memory_used_mb, :memory_used_percent, :low_memory,
		:timestamp
	)`

	_, err := r.db.NamedExec(query, report)
	return err
}

func (r *sqliteReportRepo) GetLatestByTablet(tabletID int64, onlySuccess bool) (*TabletReport, error) {
	var rep TabletReport

	query := "SELECT * FROM reports WHERE tablet_id = ?"
	if onlySuccess {
		query += " AND success = 1"
	}
	query += " ORDER BY timestamp DESC LIMIT 1"

	err := r.db.Get(&rep, query, tabletID)
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *sqliteReportRepo) GetHistory(tabletID int64, limit int) ([]TabletReport, error) {
	var history []TabletReport
	err := r.db.Select(&history, "SELECT * FROM reports WHERE tablet_id = ? ORDER BY timestamp DESC LIMIT ?", tabletID, limit)
	return history, err
}

func (r *sqliteReportRepo) Cleanup(days int) error {
	if days <= 0 {
		return nil
	}

	query := `DELETE FROM reports WHERE timestamp < DATETIME('now', '-' || ? || ' days')`
	_, err := r.db.Exec(query, days)
	return err
}
