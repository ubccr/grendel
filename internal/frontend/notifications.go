package frontend

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// To append another trigger, pass a string with a leading comma in thr form of an object. ex: ', "event": "value"'
func ToastSuccess(context *fiber.Ctx, msg string, appendTrigger string) error {
	context.Append("HX-Trigger", fmt.Sprintf(`{"toast-success": "%s"%s}`, msg, appendTrigger))
	return context.Send(nil)
}
func ToastError(context *fiber.Ctx, err error, msg string) error {
	if err != nil {
		log.Error(err)
	}
	context.Response().Header.Add("HX-Trigger", fmt.Sprintf(`{"toast-error": "%s"}`, msg))
	context.Response().Header.Add("HX-Reswap", "none")
	return context.SendString(msg)
}
