package bmc

type SystemManager interface {
	PowerCycle() error
	EnablePXE() error
}
