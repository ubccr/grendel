package frontend

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ubccr/grendel/firmware"
)

func (h *Handler) getFirmware() []string {
	fw := make([]string, 0)
	for _, i := range firmware.BuildToStringMap {
		fw = append(fw, i)
	}
	return fw
}

func (h *Handler) getBootImages() []string {
	images, _ := h.DB.BootImages()
	bootImages := make([]string, 0)
	for _, i := range images {
		bootImages = append(bootImages, i.Name)
	}
	return bootImages
}
func (h *Handler) writeEvent(severity string, f *fiber.Ctx, message string) error {
	sess, err := h.Store.Get(f)
	if err != nil {
		return err
	}
	// TODO: fix returns "%!s()" on first user registration
	user := fmt.Sprintf("%s", sess.Get("user"))
	e := EventStruct{
		Severity: severity,
		Time:     time.Now().Format("Jan 02 - 03:04:05pm"),
		User:     user,
		Message:  message,
	}

	h.Events = append(h.Events, e)

	return err
}
