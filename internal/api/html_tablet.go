package api

import (
	"fmt"
	"net/http"
	"strconv"

	"freekiosk-hub/internal/models"
	"freekiosk-hub/internal/repositories"
	"freekiosk-hub/internal/services"
	"freekiosk-hub/ui"

	"github.com/labstack/echo/v4"
)

type HtmlTabletHandler struct {
	tabletRepo repositories.TabletRepository
	reportRepo repositories.ReportRepository
	groupRepo  repositories.GroupRepository
	kService   services.KioskService
}

func NewHtmlTabletHandler(tr repositories.TabletRepository, rr repositories.ReportRepository, gr repositories.GroupRepository, ks services.KioskService) *HtmlTabletHandler {
	return &HtmlTabletHandler{tabletRepo: tr, reportRepo: rr, groupRepo: gr, kService: ks}
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

func (h *HtmlTabletHandler) HandleBeep(c echo.Context) error {
	// 1. Récupérer l'ID de la tablette depuis l'URL
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("ID de tablette invalide", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	// 2. Appeler le service Beep sur cette cible unique
	// Le service gère déjà le parallélisme et la validation du champ "executed"
	report, err := h.kService.Beep(services.Target{TabletID: id})
	if err != nil {
		// Erreur de résolution (ex: tablette non trouvée en DB)
		return ui.Toast("Erreur : "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	// 3. Générer les toasts basés sur le rapport de résultat
	// (Ici on n'a qu'un résultat car on cible un ID unique)
	for _, res := range report.Results {
		if res.Executed {
			ui.Toast(fmt.Sprintf("🔔 %s : Beep envoyé !", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.Toast(fmt.Sprintf("❌ %s : Échec du Beep ", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}

	return nil
}
