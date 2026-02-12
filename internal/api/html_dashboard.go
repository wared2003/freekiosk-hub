package api

import (
	"net/http"

	"github.com/wared2003/freekiosk-hub/internal/models"
	"github.com/wared2003/freekiosk-hub/internal/repositories"
	"github.com/wared2003/freekiosk-hub/ui"

	"github.com/labstack/echo/v4"
)

type HtmlHomeHandler struct {
	tabletRepo repositories.TabletRepository
	reportRepo repositories.ReportRepository
	groupRepo  repositories.GroupRepository
}

func NewHtmlHomeHandler(tr repositories.TabletRepository, rr repositories.ReportRepository, gr repositories.GroupRepository) *HtmlHomeHandler {
	return &HtmlHomeHandler{
		tabletRepo: tr,
		reportRepo: rr,
		groupRepo:  gr,
	}
}

func (h *HtmlHomeHandler) HandleIndex(c echo.Context) error {
	tablets, _ := h.tabletRepo.GetAll()

	var displayList []models.TabletDisplay
	for _, t := range tablets {
		report, _ := h.reportRepo.GetLatestByTablet(int64(t.ID), true)
		groups, _ := h.groupRepo.GetGroupsByTablet(int64(t.ID))
		displayList = append(displayList, models.TabletDisplay{
			Tablet:     t,
			LastReport: report,
			Groups:     groups,
		})
	}

	if c.QueryParam("refresh") == "true" {
		return ui.DashboardGrid(displayList).Render(c.Request().Context(), c.Response().Writer)
	}

	FullPage := c.Request().Header.Get("HX-Request") != "true"

	return c.Render(http.StatusOK, "", ui.Dashboard(displayList, FullPage))
}
