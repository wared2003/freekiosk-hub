package network

import (
	"log/slog"
	"net/http"

	"tailscale.com/tsnet"
)

type TailscaleNode struct {
	Server *tsnet.Server
	Client *http.Client
}

func InitTailscale(authKey string, hostname string) (*TailscaleNode, error) {
	s := &tsnet.Server{
		Hostname: hostname,
		AuthKey:  authKey,
		Logf: func(format string, args ...any) {
			slog.Debug("Tailscale", "msg", format)
		},
	}

	client := s.HTTPClient()

	slog.Info("🌐 Connexion au réseau Tailscale...", "node_name", hostname)

	return &TailscaleNode{
		Server: s,
		Client: client,
	}, nil
}

// Close ferme proprement la connexion Tailscale
func (tn *TailscaleNode) Close() {
	if tn.Server != nil {
		tn.Server.Close()
	}
}
