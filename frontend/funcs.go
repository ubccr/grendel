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
	user := sess.Get("user")
	tdClasses := "border border-neutral-300 p-1"
	msg := fmt.Sprintf(`<tr><td class="%s">%s</td><td class="%s">%s</td><td class="%s">%s</td><td class="%s">%s</td></tr>`, tdClasses, time.Now().Format("Jan 02 - 03:04:05pm"), tdClasses, user, tdClasses, severity, tdClasses, message)

	h.Events = append(h.Events, msg)

	return err
}
