package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"freekiosk-hub/internal/clients"
	"freekiosk-hub/internal/config"
	"freekiosk-hub/internal/repositories"
	"freekiosk-hub/internal/services"
	"freekiosk-hub/internal/sse"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// ApiServer centralise les dépendances pour le routage
type ApiServer struct {
	Echo        *echo.Echo
	DB          *sql.DB
	TabletRepo  repositories.TabletRepository
	ReportRepo  repositories.ReportRepository
	GroupRepo   repositories.GroupRepository
	MonitorSvc  services.MonitorService
	KioskClient clients.KioskClient
	cfg         config.Config
}

// NewRouter initialise le serveur, les handlers et les routes
func NewRouter(e *echo.Echo, db *sql.DB,
	tr repositories.TabletRepository,
	rr repositories.ReportRepository,
	gr repositories.GroupRepository,
	ms services.MonitorService,
	ks clients.KioskClient,
	cfg config.Config,

) *ApiServer {
	s := &ApiServer{
		Echo:        e,
		DB:          db,
		TabletRepo:  tr,
		ReportRepo:  rr,
		GroupRepo:   gr,
		MonitorSvc:  ms,
		KioskClient: ks,
		cfg:         cfg,
	}

	s.setupMiddlewares()
	s.setupRoutes()

	return s
}

func (s *ApiServer) setupMiddlewares() {
	// Nouveau RequestLogger : Plus propre et structuré
	s.Echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		LogRemoteIP: true,
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "/sse")
		},
		HandleError: true, // Pour que les erreurs passent aussi par ici
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				slog.Error("HTTP Request Error",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency,
					"remote_ip", v.RemoteIP,
					"error", v.Error,
				)
			} else {
				slog.Info("HTTP Request",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency,
					"remote_ip", v.RemoteIP,
				)
			}
			return nil
		},
	}))

	s.Echo.Use(middleware.Recover())
	s.Echo.Static("/static", "static")
}

func (s *ApiServer) setupRoutes() {

	kService := services.NewKioskService(s.TabletRepo, s.GroupRepo, s.KioskClient, s.cfg.KioskPort)

	homeH := NewHtmlHomeHandler(s.TabletRepo, s.ReportRepo, s.GroupRepo)
	tabletH := NewHtmlTabletHandler(s.TabletRepo, s.ReportRepo, s.GroupRepo, kService)
	groupH := NewGroupHandler(s.GroupRepo)

	systemJsonH := NewSystemJSONHandler(s.DB)

	// --- 2. ROUTES PUBLIQUES / SYSTÈME ---
	s.Echo.GET("/health", systemJsonH.HandleHealthCheck)

	s.Echo.GET("/", homeH.HandleIndex)
	tablets := s.Echo.Group("/tablets")
	{
		tablets.GET("/:id", tabletH.HandleDetails)
		tablets.GET("/:id/groups-selection", groupH.HandleTabletGroupsSelection)
		tablets.POST("/:tabletID/groups/:groupID/toggle", groupH.HandleToggleGroup)

		//commands
		tablets.POST("/:id/command/beep", tabletH.HandleBeep)
	}

	groupRoutes := s.Echo.Group("/groups")
	{
		groupRoutes.GET("", groupH.HandleGroups)
		groupRoutes.GET("/new", groupH.HandleNewGroup)
		groupRoutes.GET("/edit/:id", groupH.HandleEditGroup)
		groupRoutes.POST("/save", groupH.HandleSaveGroup)
		groupRoutes.DELETE("/:id", groupH.HandleDeleteGroup)
	}

	// s.Echo.GET("/admin/import", adminPageH.HandleImportPage)

	// // --- 4. ROUTES API (JSON) ---
	// // On groupe les routes API sous /api/v1
	// apiV1 := s.Echo.Group("/api/v1")

	// // Si une clé API est configurée, on pourrait ajouter un middleware ici
	// // apiV1.Use(CustomApiKeyMiddleware(s.ApiKey))

	// apiV1.GET("/tablets", tabletJsonH.HandleListTablets)
	// apiV1.POST("/tablets/import", tabletJsonH.HandleBulkImport) // Pour tes 500 tablettes
	// apiV1.POST("/tablets/:ip/scan", tabletJsonH.HandleManualScan)

	//sse
	s.Echo.GET("/sse/global", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("X-Accel-Buffering", "no")
		fmt.Fprintf(c.Response().Writer, "data: connected\n\n")
		c.Response().Flush()

		ch := sse.Instance.SubscribeGlobal()
		defer sse.Instance.Unsubscribe(ch, 0)
		for {
			select {
			case <-ch:
				// On envoie l'event "refresh"
				fmt.Fprintf(c.Response().Writer, "event: update\ndata: \n\n")
				c.Response().Flush()
			case <-c.Request().Context().Done():
				return nil
			}
		}
	})

	s.Echo.GET("/sse/tablet/:id", func(c echo.Context) error {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("X-Accel-Buffering", "no")
		fmt.Fprintf(c.Response().Writer, "data: connected\n\n")
		c.Response().Flush()

		ch := sse.Instance.SubscribeTablet(id)
		defer sse.Instance.Unsubscribe(ch, id)

		for {
			select {
			case <-ch:
				fmt.Fprintf(c.Response().Writer, "event: update\ndata: \n\n")
				c.Response().Flush()
			case <-c.Request().Context().Done():
				return nil
			}
		}
	})
}
