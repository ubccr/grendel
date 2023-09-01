package frontend

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ToastHtml(m, t string) string {
	if t == "success" {
		return "<div class='rounded-lg border border-neutral-400 bg-white p-4 shadow-lg z-100'><h1 class='text-green-600'>" + m + "</h1></div>"
	} else {
		return "<div class='rounded-lg border border-neutral-400 bg-white p-4 shadow-lg z-100'><h1 class='text-red-600'>" + m + "</h1></div>"
	}
}

func ToastSuccess(context echo.Context, msg string) error {
	context.Response().Header().Add("HX-Trigger", fmt.Sprintf(`{"toast-success": "%s"}`, msg))
	return context.NoContent(http.StatusNoContent)
}
func ToastError(context echo.Context, err error, msg string) error {
	log.Error(err)
	context.Response().Header().Add("HX-Trigger", fmt.Sprintf(`{"toast-error": "%s"}`, msg))
	return context.String(http.StatusBadRequest, msg)
}
