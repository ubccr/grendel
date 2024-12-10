package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/pkg/model"
)

func (h *Handler) Restore(c echo.Context) error {
	var dump model.DataDump

	if !strings.HasPrefix(c.Request().Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content type")
	}

	if err := c.Bind(&dump); err != nil {
		return err
	}

	log.Infof("Attempting to restore %d users, %d hosts, and %d boot images", len(dump.Users), len(dump.Hosts), len(dump.Images))

	err := h.DB.RestoreFrom(dump)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to restore database").SetInternal(err)
	}

	log.Infof("Database restored successfully")

	res := map[string]interface{}{
		"ok": true,
	}
	return c.JSON(http.StatusCreated, res)
}
