package frontend

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func ToastSuccess(context *fiber.Ctx, msg string) error {
	context.Response().Header.Add("HX-Trigger", fmt.Sprintf(`{"toast-success": "%s"}`, msg))
	return context.Send(nil)
}
func ToastError(context *fiber.Ctx, err error, msg string) error {
	log.Error(err)
	context.Response().Header.Add("HX-Trigger", fmt.Sprintf(`{"toast-error": "%s"}`, msg))
	context.Response().Header.Add("HX-Reswap", "none")
	return context.SendString(msg)
}
