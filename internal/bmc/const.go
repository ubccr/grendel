package bmc

import "github.com/spf13/viper"

const (
	delay  = 1
	fanout = 5
)

func init() {
	viper.SetDefault("bmc.delay", delay)
	viper.SetDefault("bmc.fanout", fanout)
}
