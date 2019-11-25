package model

type BootSpec struct {
	Name    string
	Kernel  string
	Initrd  []string
	Cmdline string
	Message string
}
