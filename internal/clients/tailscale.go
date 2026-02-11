package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TailscaleDevice représente un appareil retourné par l'API Tailscale
type TailscaleDevice struct {
	ID        string   `json:"id"`
	Hostname  string   `json:"hostname"`
	Addresses []string `json:"addresses"` // Contiendra l'IP 100.x.y.z
	Tags      []string `json:"tags"`
}

type TailscaleClient struct {
	apiKey     string
	tailnet    string
	httpClient *http.Client
}

func NewTailscaleClient(apiKey, tailnet string) *TailscaleClient {
	return &TailscaleClient{
		apiKey:  apiKey,
		tailnet: tailnet,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetKiosks récupère la liste des machines sur le Tailnet
func (c *TailscaleClient) GetKiosks() ([]TailscaleDevice, error) {
	url := fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/devices", c.tailnet)

	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.apiKey, "") // L'API Key s'utilise comme username dans le BasicAuth

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API Tailscale: %d", resp.StatusCode)
	}

	var result struct {
		Devices []TailscaleDevice `json:"devices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Devices, nil
}
