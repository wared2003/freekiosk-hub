package api

import (
	"net/http"

	"freekiosk-hub/internal/models"
	"freekiosk-hub/internal/repositories"
	"freekiosk-hub/ui"

	"github.com/labstack/echo/v4"
)

type HtmlHomeHandler struct {
	tabletRepo repositories.TabletRepository
	reportRepo repositories.ReportRepository
}

func NewHtmlHomeHandler(tr repositories.TabletRepository, rr repositories.ReportRepository) *HtmlHomeHandler {
	return &HtmlHomeHandler{
		tabletRepo: tr,
		reportRepo: rr,
	}
}

func (h *HtmlHomeHandler) HandleIndex(c echo.Context) error {
	tablets, _ := h.tabletRepo.GetAll()

	var displayList []models.TabletDisplay
	for _, t := range tablets {
		report, _ := h.reportRepo.GetLatestByTablet(int64(t.ID), true)
		displayList = append(displayList, models.TabletDisplay{
			Tablet:     t,
			LastReport: report,
		})
	}

	// Si c'est du HTMX, on rend le template Dashboard seul
	// Sinon, on rend le template Dashboard "complet" (le template gère le layout)
	FullPage := c.Request().Header.Get("HX-Request") != "true"

	return c.Render(http.StatusOK, "", ui.Dashboard(displayList, FullPage))
}
