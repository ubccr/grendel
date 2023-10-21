package frontend

import "github.com/ubccr/grendel/firmware"

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
