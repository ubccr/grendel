package model

type BootImage struct {
	ID          uint64   `json:"id" badgerhold:"key"`
	Name        string   `json:"name"`
	KernelPath  string   `json:"kernel"`
	InitrdPaths []string `json:"initrd"`
	LiveImage   string   `json:"liveimg"`
	RootFS      string   `json:"rootfs"`
	InstallRepo string   `json:"install_repo"`
	CommandLine string   `json:"cmdline"`
}
