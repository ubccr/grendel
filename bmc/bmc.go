package bmc

import "github.com/stmcginnis/gofish"

// type BmcActions interface {
// 	PowerCycle() error
// 	PowerOn() error
// 	PowerOff() error
// }

type Redfish2 struct {
	config  gofish.ClientConfig
	client  *gofish.APIClient
	service *gofish.Service
}
