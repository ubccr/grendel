package model

import (
	"github.com/segmentio/ksuid"
)

type BootImageList []*BootImage

type BootImage struct {
	ID          ksuid.KSUID `json:"id"`
	Name        string      `json:"name"`
	KernelPath  string      `json:"kernel"`
	InitrdPaths []string    `json:"initrd"`
	LiveImage   string      `json:"liveimg"`
	InstallRepo string      `json:"install_repo"`
	CommandLine string      `json:"cmdline"`
}

func NewBootImageList() BootImageList {
	return make(BootImageList, 0)
}
