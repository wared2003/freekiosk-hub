package api

import (
	"net/http"
	"strconv"

	"freekiosk-hub/internal/models"
	"freekiosk-hub/internal/repositories"
	"freekiosk-hub/ui"

	"github.com/labstack/echo/v4"
)

type HtmlTabletHandler struct {
	tabletRepo repositories.TabletRepository
	reportRepo repositories.ReportRepository
	groupRepo  repositories.GroupRepository
}

func NewHtmlTabletHandler(tr repositories.TabletRepository, rr repositories.ReportRepository, gr repositories.GroupRepository) *HtmlTabletHandler {
	return &HtmlTabletHandler{tabletRepo: tr, reportRepo: rr, groupRepo: gr}
}

func (h *HtmlTabletHandler) HandleDetails(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "ID invalide")
	}

	tablet, err := h.tabletRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Tablette non trouvée")
	}

	lastReport, _ := h.reportRepo.GetLatestByTablet(id, true)

	history, _ := h.reportRepo.GetHistory(id, 30)

	groups, _ := h.groupRepo.GetGroupsByTablet(id)

	td := models.TabletDisplay{
		Tablet:     *tablet,
		LastReport: lastReport,
		Groups:     groups,
	}

	if c.Request().Header.Get("HX-Request") != "true" {
		return c.Render(http.StatusOK, "", ui.TabletDetails(&td, history, true))
	}

	// 2. Si c'est un refresh auto du SSE (on ajoute ?refresh=true dans le hx-get du template)
	if c.QueryParam("refresh") == "true" {
		return c.Render(http.StatusOK, "", ui.TabletUIInner(&td, history))
	}

	return c.Render(http.StatusOK, "", ui.TabletDetails(&td, history, false))
}
