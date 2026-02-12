package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"freekiosk-hub/internal/repositories"
	"freekiosk-hub/ui"

	"github.com/labstack/echo/v4"
)

type GroupHandler struct {
	groupRepo repositories.GroupRepository
}

func NewGroupHandler(gr repositories.GroupRepository) *GroupHandler {
	return &GroupHandler{groupRepo: gr}
}

// GET /groups
func (h *GroupHandler) HandleGroups(c echo.Context) error {
	groups, err := h.groupRepo.GetAll()
	if err != nil {
		slog.Error("database error: failed to fetch groups", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	tableData := make(map[int64][]repositories.Tablet)
	for _, g := range groups {
		tablets, err := h.groupRepo.GetTabletsByGroup(g.ID)
		if err != nil {
			slog.Warn("data integrity: could not load tablets for group", "group_id", g.ID, "err", err)
			tableData[g.ID] = []repositories.Tablet{}
			continue
		}
		tableData[g.ID] = tablets
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, "", ui.GroupsContent(groups, tableData))
	}
	return c.Render(http.StatusOK, "", ui.GroupsPage(groups, tableData))
}

// GET /groups/new
func (h *GroupHandler) HandleNewGroup(c echo.Context) error {
	return c.Render(http.StatusOK, "", ui.GroupFormModal(&repositories.Group{
		Color: "#6366f1",
	}))
}

// GET /groups/edit/:id
func (h *GroupHandler) HandleEditGroup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		slog.Warn("invalid request: malformed group id", "id_param", c.Param("id"))
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	group, err := h.groupRepo.GetByID(id)
	if err != nil {
		slog.Error("database error: group not found", "id", id, "err", err)
		return c.String(http.StatusNotFound, "Group not found")
	}

	return c.Render(http.StatusOK, "", ui.GroupFormModal(group))
}

// POST /groups/save
func (h *GroupHandler) HandleSaveGroup(c echo.Context) error {
	id, _ := strconv.ParseInt(c.FormValue("id"), 10, 64)

	group := &repositories.Group{
		ID:          id,
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Color:       c.FormValue("color"),
	}

	var err error
	if group.ID == 0 {
		_, err = h.groupRepo.Create(group)
		slog.Info("resource created: new group added", "name", group.Name)
	} else {
		err = h.groupRepo.Update(group)
		slog.Info("resource updated: group modified", "id", group.ID, "name", group.Name)
	}

	if err != nil {
		slog.Error("database error: failed to save group", "err", err, "group_name", group.Name)
		return c.String(http.StatusInternalServerError, "Failed to save group")
	}

	return h.HandleGroups(c)
}

// DELETE /groups/:id
func (h *GroupHandler) HandleDeleteGroup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	if err := h.groupRepo.Delete(id); err != nil {
		slog.Error("database error: failed to delete group", "id", id, "err", err)
		return c.String(http.StatusInternalServerError, "Deletion failed")
	}

	slog.Info("resource deleted: group removed", "id", id)
	return c.NoContent(http.StatusOK)
}
