package frontend

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ubccr/grendel/internal/bmc"
	"github.com/ubccr/grendel/internal/firmware"
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

	log.Debugf("User: %s - Severity: %s - Message: %s", e.User, e.Severity, e.Message)

	h.Events = append([]EventStruct{e}, h.Events...)

	if len(h.Events) > 50 {
		h.Events = h.Events[:50]
	}

	return err
}

func (h *Handler) writeJobEvent(f *fiber.Ctx, message string, jobMessages []bmc.JobMessage) error {
	sess, err := h.Store.Get(f)
	if err != nil {
		return err
	}
	// TODO: fix returns "%!s()" on first user registration
	user := fmt.Sprintf("%s", sess.Get("user"))

	// TODO: loop over jobMessages and tally success vs failure to determine severity
	e := EventStruct{
		Severity:    "info",
		Time:        time.Now().Format("Jan 02 - 03:04:05pm"),
		User:        user,
		Message:     message,
		JobMessages: jobMessages,
	}

	log.Debugf("User: %s - Severity: %s - Message: %s", e.User, e.Severity, e.Message)

	h.Events = append([]EventStruct{e}, h.Events...)

	if len(h.Events) > 50 {
		h.Events = h.Events[:50]
	}

	return err
}
