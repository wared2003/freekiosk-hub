package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type SystemJSONHandler struct {
	db *sql.DB
}

func NewSystemJSONHandler(db *sql.DB) *SystemJSONHandler {
	return &SystemJSONHandler{
		db: db,
	}
}

// HealthResponse définit la structure de sortie du healthcheck
type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// HandleHealthCheck vérifie l'état de santé du Hub
func (h *SystemJSONHandler) HandleHealthCheck(c echo.Context) error {
	dbStatus := "connected"

	// On tente un Ping sur la base SQLite
	if err := h.db.Ping(); err != nil {
		dbStatus = "disconnected"
		return c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:    "unhealthy",
			Database:  dbStatus,
			Timestamp: time.Now(),
			Version:   "0.0.2",
		})
	}

	return c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Database:  dbStatus,
		Timestamp: time.Now(),
		Version:   "0.0.2",
	})
}
