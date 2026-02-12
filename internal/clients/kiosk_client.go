package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"freekiosk-hub/internal/repositories"
)

type KioskClient interface {
	// Statut & Monitoring
	FetchStatus(ip string) (*repositories.TabletReport, error)

	// Affichage & UI
	SetBrightness(ip string, value int) error
	SetVolume(ip string, value int) error
	SetScreen(ip string, on bool) error
	SetScreensaver(ip string, active bool) error
	ShowToast(ip string, text string) error

	// Navigation & Webview
	Navigate(ip string, url string) error
	NavigateAlias(ip string, url string) error
	Reload(ip string) error
	ClearCache(ip string) error
	ExecuteJS(ip string, code string) error
	SetRotation(ip string, start bool) error

	// Médias & Interaction
	Speak(ip string, text string) error
	PlayAudio(ip string, url string, loop bool, volume int) error
	StopAudio(ip string) error
	Beep(ip string) error

	// Système & Apps
	Wake(ip string) error
	Reboot(ip string) error
	LaunchApp(ip string, packageName string) error

	// Caméra
	TakePhoto(ip string, camera string, quality int) ([]byte, error)

	// Contrôle à distance
	SendRemoteCommand(ip string, action string) error
}

type httpClientImpl struct {
	httpClient *http.Client
}

func NewKioskClient(client *http.Client) KioskClient {
	return &httpClientImpl{
		httpClient: client,
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

	resp, err := c.httpClient.Get(url)
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

type CommandResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Executed bool   `json:"executed"`
		Command  string `json:"command"`
	} `json:"data"`
}

func (c *httpClientImpl) postJSON(ip, path string, payload interface{}) error {
	url := fmt.Sprintf("http://%s%s", ip, path)
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// On décode la réponse systématiquement
	var cr CommandResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return fmt.Errorf("failed to decode kiosk response: %w", err)
	}

	// C'est ici qu'on vérifie tes deux drapeaux
	if !cr.Success {
		return fmt.Errorf("kiosk returned success=false")
	}
	if !cr.Data.Executed {
		return fmt.Errorf("kiosk failed to execute %s", cr.Data.Command)
	}

	return nil
}
func (c *httpClientImpl) SetBrightness(ip string, value int) error {
	return c.postJSON(ip, "/api/brightness", map[string]int{"value": value})
}

func (c *httpClientImpl) SetVolume(ip string, value int) error {
	return c.postJSON(ip, "/api/volume", map[string]int{"value": value})
}

func (c *httpClientImpl) Navigate(ip string, url string) error {
	return c.postJSON(ip, "/api/url", map[string]string{"url": url})
}

func (c *httpClientImpl) Speak(ip string, text string) error {
	return c.postJSON(ip, "/api/tts", map[string]string{"text": text})
}

func (c *httpClientImpl) ShowToast(ip string, text string) error {
	return c.postJSON(ip, "/api/toast", map[string]string{"text": text})
}

func (c *httpClientImpl) SetScreen(ip string, on bool) error {
	path := "/api/screen/off"
	if on {
		path = "/api/screen/on"
	}
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s%s", ip, path), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) SetScreensaver(ip string, active bool) error {
	path := "/api/screensaver/off"
	if active {
		path = "/api/screensaver/on"
	}
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s%s", ip, path), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) Reload(ip string) error {
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s/api/reload", ip), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) Wake(ip string) error {
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s/api/wake", ip), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) Reboot(ip string) error {
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s/api/reboot", ip), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) ClearCache(ip string) error {
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s/api/clearCache", ip), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) LaunchApp(ip string, packageName string) error {
	return c.postJSON(ip, "/api/app/launch", map[string]string{"package": packageName})
}

func (c *httpClientImpl) ExecuteJS(ip string, code string) error {
	return c.postJSON(ip, "/api/js", map[string]string{"code": code})
}

func (c *httpClientImpl) TakePhoto(ip string, camera string, quality int) ([]byte, error) {
	url := fmt.Sprintf("http://%s/api/camera/photo?camera=%s&quality=%d", ip, camera, quality)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *httpClientImpl) PlayAudio(ip string, url string, loop bool, volume int) error {
	payload := map[string]interface{}{"url": url, "loop": loop, "volume": volume}
	return c.postJSON(ip, "/api/audio/play", payload)
}

func (c *httpClientImpl) StopAudio(ip string) error {
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s/api/audio/stop", ip), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) Beep(ip string) error {
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s/api/audio/beep", ip), "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) SendRemoteCommand(ip string, action string) error {
	url := fmt.Sprintf("http://%s/api/remote/%s", ip, action)
	resp, err := c.httpClient.Post(url, "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) SetRotation(ip string, start bool) error {
	path := "/api/rotation/stop"
	if start {
		path = "/api/rotation/start"
	}
	url := fmt.Sprintf("http://%s%s", ip, path)
	resp, err := c.httpClient.Post(url, "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (c *httpClientImpl) NavigateAlias(ip string, url string) error {
	return c.postJSON(ip, "/api/navigate", map[string]string{"url": url})
}

func (c *httpClientImpl) WakeFromScreensaver(ip string) error {
	url := fmt.Sprintf("http://%s/api/wake", ip)
	resp, err := c.httpClient.Post(url, "", nil)
	if err == nil {
		resp.Body.Close()
	}
	return err
}
