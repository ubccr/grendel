package bmc

type SystemManager interface {
	PowerCycle() error
	EnablePXE() error
	Logout()
	GetSystem() (*System, error)
}

type System struct {
	Name           string   `json:"name"`
	BIOSVersion    string   `json:"bios_version"`
	SerialNumber   string   `json:"serial_number"`
	Manufacturer   string   `json:"manufacturer"`
	PowerStatus    string   `json:"power_status"`
	Health         string   `json:"health"`
	TotalMemory    float32  `json:"total_memory"`
	ProcessorCount int      `json:"processor_count"`
	BootNext       string   `json:"boot_next"`
	BootOrder      []string `json:"boot_order"`
}
