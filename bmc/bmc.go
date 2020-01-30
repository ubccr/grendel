// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

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
