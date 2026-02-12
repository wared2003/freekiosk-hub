package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/wared2003/freekiosk-hub/internal/models"
	"github.com/wared2003/freekiosk-hub/internal/repositories"
	"github.com/wared2003/freekiosk-hub/internal/services"
	"github.com/wared2003/freekiosk-hub/ui"

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
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("invalid tablet id", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	report, err := h.kService.Beep(services.Target{TabletID: id})
	if err != nil {
		return ui.Toast("error : "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	for _, res := range report.Results {
		if res.Executed {
			ui.Toast(fmt.Sprintf("🔔 %s : Beep Send !", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.Toast(fmt.Sprintf("❌ %s : Error sending Beep ", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}

	return nil
}

func (h *HtmlTabletHandler) HandleReload(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("invalid tablet id ", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	report, err := h.kService.Reload(services.Target{TabletID: id})
	if err != nil {
		return ui.Toast("Erreur : "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	for _, res := range report.Results {
		if res.Executed {
			ui.Toast(fmt.Sprintf("🔄 %s : Reloading...", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.Toast(fmt.Sprintf("❌ %s : error reloading", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}

	return nil
}

func (h *HtmlTabletHandler) HandleReboot(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("invalid tablet id ", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	report, err := h.kService.Reboot(services.Target{TabletID: id})
	if err != nil {
		return ui.Toast("Erreur : "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	for _, res := range report.Results {
		if res.Executed {
			ui.Toast(fmt.Sprintf("🔄 %s : Rebooting", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.Toast(fmt.Sprintf("❌ %s : error reboot failed", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}

	return nil
}

func (h *HtmlTabletHandler) HandleNavigateModal(c echo.Context) error {
	idParam := c.Param("id")
	id, _ := strconv.ParseInt(idParam, 10, 64)

	// On peut optionnellement récupérer l'URL actuelle depuis la DB/Cache
	// pour pré-remplir l'input
	currentURL := ""

	return ui.NavigateModal(id, currentURL).Render(c.Request().Context(), c.Response().Writer)
}

func (h *HtmlTabletHandler) HandleNavigate(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("Invalid tablet ID", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	newURL := c.FormValue("url")
	if newURL == "" {
		return ui.Toast("URL cannot be empty", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	parsedURL, err := url.ParseRequestURI(newURL)
	if err != nil {
		return ui.Toast("Invalid URL format", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ui.Toast("Forbidden protocol: Use HTTP or HTTPS", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	report, err := h.kService.Navigate(services.Target{TabletID: id}, parsedURL.String())
	if err != nil {
		return ui.Toast("Error: "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	for _, res := range report.Results {
		if res.Executed {
			ui.Toast(fmt.Sprintf("🌐 %s: URL updated!", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.Toast(fmt.Sprintf("❌ %s: Update failed", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}
	c.Response().Header().Set("HX-Trigger", "update")
	return nil
}

func (h *HtmlTabletHandler) HandleWakeUp(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("invalid tablet id ", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	report, err := h.kService.Wake(services.Target{TabletID: id})
	if err != nil {
		return ui.Toast("Error : "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	for _, res := range report.Results {
		if res.Executed {
			ui.Toast(fmt.Sprintf("⏰ %s : Waked up", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.Toast(fmt.Sprintf("❌ %s : error waking up", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}

	return nil
}

func (h *HtmlTabletHandler) HandleScreenStatus(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("Invalid tablet ID", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	statusRaw := c.FormValue("status")
	var shouldBeOn bool
	switch statusRaw {
	case "true", "on":
		shouldBeOn = true
	case "false", "off":
		shouldBeOn = false
	default:
		return ui.Toast("err: invalid request", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	report, err := h.kService.SetScreen(services.Target{TabletID: id}, shouldBeOn)
	if err != nil {
		ui.ScreenStatusBox(!shouldBeOn, id).Render(c.Request().Context(), c.Response().Writer)
		return ui.Toast("Error: "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	for _, res := range report.Results {
		if res.Executed {
			ui.ScreenStatusBox(shouldBeOn, id).Render(c.Request().Context(), c.Response().Writer)
			ui.Toast(fmt.Sprintf("%s :screen command send", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.ScreenStatusBox(!shouldBeOn, id).Render(c.Request().Context(), c.Response().Writer)
			ui.Toast(fmt.Sprintf("❌ %s: send screen command failed", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}
	return nil
}

func (h *HtmlTabletHandler) HandleScreenSaver(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ui.Toast("Invalid tablet ID", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	statusRaw := c.FormValue("status")
	var shouldBeOn bool
	switch statusRaw {
	case "true", "on":
		shouldBeOn = true
	case "false", "off":
		shouldBeOn = false
	default:
		return ui.Toast("err: invalid request", "error").Render(c.Request().Context(), c.Response().Writer)
	}

	report, err := h.kService.SetScreensaver(services.Target{TabletID: id}, shouldBeOn)
	if err != nil {
		ui.ScreensaverStatusBox(!shouldBeOn, id).Render(c.Request().Context(), c.Response().Writer)
		return ui.Toast("Error: "+err.Error(), "error").Render(c.Request().Context(), c.Response().Writer)
	}

	for _, res := range report.Results {
		if res.Executed {
			ui.ScreensaverStatusBox(shouldBeOn, id).Render(c.Request().Context(), c.Response().Writer)
			ui.Toast(fmt.Sprintf("%s :screensaver command send", res.Name), "success").Render(c.Request().Context(), c.Response().Writer)
		} else {
			ui.ScreensaverStatusBox(!shouldBeOn, id).Render(c.Request().Context(), c.Response().Writer)
			ui.Toast(fmt.Sprintf("❌ %s: send screensaver command failed", res.Name), "error").Render(c.Request().Context(), c.Response().Writer)
		}
	}
	return nil
}
