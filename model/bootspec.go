package model

type BootSpec struct {
	Name    string
	Kernel  []byte
	Initrd  [][]byte
	Cmdline string
	Message string
}
