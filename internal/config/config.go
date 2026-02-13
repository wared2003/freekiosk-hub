package config

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	DBPath        string
	PollInterval  time.Duration
	MaxWorkers    int
	TSAuthKey     string
	LogLevel      string
	KioskPort     string
	RetentionDays int
	KioskApiKey   string
	MediaDir      string
	BaseURL       string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ Aucun fichier .env trouvé, utilisation des variables système ou par défaut")
	}

	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", "8081"),
		DBPath:        getEnv("DB_PATH", "freekiosk.db"),
		PollInterval:  parseDuration(getEnv("POLL_INTERVAL", "30s")),
		MaxWorkers:    parseInt(getEnv("MAX_WORKERS", "5")),
		TSAuthKey:     os.Getenv("TS_AUTHKEY"),
		LogLevel:      getEnv("LOG_LEVEL", "INFO"),
		KioskPort:     getEnv("KIOSK_PORT", "8080"),
		RetentionDays: parseInt(getEnv("RETENTION_DAYS", "31")),
		KioskApiKey:   getEnv("KIOSK_API_KEY", ""),
		MediaDir:      getEnv("MEDIA_DIR", "media"),
		BaseURL:       getEnv("BASE_URL", "localhost:8081"),
	}

	initLogger(cfg.LogLevel)

	return cfg
}

func initLogger(level string) {
	var slogLevel slog.Level

	switch level {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "WARN":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		slog.Warn("Intervalle de temps invalide, retour à 30s", "valeur", s)
		return 30 * time.Second
	}
	return d
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		slog.Warn("Nombre entier invalide, retour à 5", "valeur", s)
		return 5
	}
	return i
}
