package api

import (
	"io"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if component, ok := data.(templ.Component); ok {
		return component.Render(c.Request().Context(), w)
	}
	return nil
}
