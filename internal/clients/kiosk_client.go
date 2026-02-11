package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"freekiosk-hub/internal/repositories"
)

type KioskClient interface {
	FetchStatus(ip string) (*repositories.TabletReport, error)
}

type httpClientImpl struct {
	httpClient *http.Client
	apiKey     string
}

func NewKioskClient(client *http.Client, apiKey string) KioskClient {
	return &httpClientImpl{
		httpClient: client,
		apiKey:     apiKey,
	}
}

type kioskResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Battery struct {
			Level    int    `json:"level"`
			Charging bool   `json:"charging"`
			Plugged  string `json:"plugged"`
		} `json:"battery"`
		Screen struct {
			On                bool `json:"on"`
			Brightness        int  `json:"brightness"`
			ScreensaverActive bool `json:"screensaverActive"`
		} `json:"screen"`
		Audio struct {
			Volume int `json:"volume"`
		} `json:"audio"`
		Webview struct {
			CurrentUrl string `json:"currentUrl"`
			CanGoBack  bool   `json:"canGoBack"`
			Loading    bool   `json:"loading"`
		} `json:"webview"`
		Device struct {
			Ip            string `json:"ip"`
			Hostname      string `json:"hostname"`
			Version       string `json:"version"`
			IsDeviceOwner bool   `json:"isDeviceOwner"`
			KioskMode     bool   `json:"kioskMode"`
		} `json:"device"`
		Wifi struct {
			Ssid           string `json:"ssid"`
			SignalStrength int    `json:"signalStrength"`
			SignalLevel    int    `json:"signalLevel"`
			Connected      bool   `json:"connected"`
			LinkSpeed      int    `json:"linkSpeed"`
			Frequency      int    `json:"frequency"`
		} `json:"wifi"`
		Rotation struct {
			Enabled      bool `json:"enabled"`
			Interval     int  `json:"interval"`
			CurrentIndex int  `json:"currentIndex"`
		} `json:"rotation"`
		Sensors struct {
			Light         float64 `json:"light"`
			Proximity     float64 `json:"proximity"`
			Accelerometer struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
				Z float64 `json:"z"`
			} `json:"accelerometer"`
		} `json:"sensors"`
		AutoBrightness struct {
			Enabled           bool    `json:"enabled"`
			Min               float64 `json:"min"`
			Max               float64 `json:"max"`
			CurrentLightLevel float64 `json:"currentLightLevel"`
		} `json:"autoBrightness"`
		Storage struct {
			TotalMB     int `json:"totalMB"`
			AvailableMB int `json:"availableMB"`
			UsedMB      int `json:"usedMB"`
			UsedPercent int `json:"usedPercent"`
		} `json:"storage"`
		Memory struct {
			TotalMB     int  `json:"totalMB"`
			AvailableMB int  `json:"availableMB"`
			UsedMB      int  `json:"usedMB"`
			UsedPercent int  `json:"usedPercent"`
			LowMemory   bool `json:"lowMemory"`
		} `json:"memory"`
	} `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

func (c *httpClientImpl) FetchStatus(ip string) (*repositories.TabletReport, error) {
	url := fmt.Sprintf("http://%s/api/status", ip)

	// Default failure report
	failReport := &repositories.TabletReport{
		Success:   false,
		Timestamp: time.Now(),
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return failReport, err
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return failReport, fmt.Errorf("failed to reach kiosk at %s: %w", ip, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return failReport, fmt.Errorf("kiosk %s returned HTTP error: %d", ip, resp.StatusCode)
	}

	var kr kioskResponse
	if err := json.NewDecoder(resp.Body).Decode(&kr); err != nil {
		return failReport, fmt.Errorf("failed to decode kiosk data (%s): %w", ip, err)
	}

	if !kr.Success {
		return failReport, fmt.Errorf("kiosk %s reported success=false", ip)
	}

	// Full success mapping
	return &repositories.TabletReport{
		Success:               true,
		BatteryLevel:          kr.Data.Battery.Level,
		BatteryCharging:       kr.Data.Battery.Charging,
		BatteryPlugged:        kr.Data.Battery.Plugged,
		ScreenOn:              kr.Data.Screen.On,
		ScreenBrightness:      kr.Data.Screen.Brightness,
		ScreensaverActive:     kr.Data.Screen.ScreensaverActive,
		AudioVolume:           kr.Data.Audio.Volume,
		CurrentURL:            kr.Data.Webview.CurrentUrl,
		WebviewCanGoBack:      kr.Data.Webview.CanGoBack,
		WebviewLoading:        kr.Data.Webview.Loading,
		DeviceIP:              kr.Data.Device.Ip,
		DeviceHostname:        kr.Data.Device.Hostname,
		DeviceVersion:         kr.Data.Device.Version,
		IsDeviceOwner:         kr.Data.Device.IsDeviceOwner,
		KioskMode:             kr.Data.Device.KioskMode,
		WifiSSID:              kr.Data.Wifi.Ssid,
		WifiSignalStrength:    kr.Data.Wifi.SignalStrength,
		WifiSignalLevel:       kr.Data.Wifi.SignalLevel,
		WifiConnected:         kr.Data.Wifi.Connected,
		WifiLinkSpeed:         kr.Data.Wifi.LinkSpeed,
		WifiFrequency:         kr.Data.Wifi.Frequency,
		RotationEnabled:       kr.Data.Rotation.Enabled,
		RotationInterval:      kr.Data.Rotation.Interval,
		RotationCurrentIndex:  kr.Data.Rotation.CurrentIndex,
		LightLevel:            kr.Data.Sensors.Light,
		Proximity:             kr.Data.Sensors.Proximity,
		AccelX:                kr.Data.Sensors.Accelerometer.X,
		AccelY:                kr.Data.Sensors.Accelerometer.Y,
		AccelZ:                kr.Data.Sensors.Accelerometer.Z,
		AutoBrightnessEnabled: kr.Data.AutoBrightness.Enabled,
		AutoBrightnessMin:     kr.Data.AutoBrightness.Min,
		AutoBrightnessMax:     kr.Data.AutoBrightness.Max,
		AutoBrightnessCurrent: kr.Data.AutoBrightness.CurrentLightLevel,
		StorageTotal:          kr.Data.Storage.TotalMB,
		StorageAvailable:      kr.Data.Storage.AvailableMB,
		StorageUsed:           kr.Data.Storage.UsedMB,
		StorageUsedPct:        kr.Data.Storage.UsedPercent,
		MemoryTotal:           kr.Data.Memory.TotalMB,
		MemoryAvailable:       kr.Data.Memory.AvailableMB,
		MemoryUsed:            kr.Data.Memory.UsedMB,
		MemoryUsedPct:         kr.Data.Memory.UsedPercent,
		LowMemory:             kr.Data.Memory.LowMemory,
		Timestamp:             time.Now(),
	}, nil
}
