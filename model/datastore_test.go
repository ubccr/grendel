package model

import (
	"net"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/stretchr/testify/assert"
)

var NetInterfaceFactory = factory.NewFactory(
	&NetInterface{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(2, 10)), nil
}).Attr("FQDN", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).Attr("MAC", func(args factory.Args) (interface{}, error) {
	return net.ParseMAC(randomdata.MacAddress())
}).Attr("IP", func(args factory.Args) (interface{}, error) {
	return net.ParseIP(randomdata.IpV4Address()), nil
})

var HostFactory = factory.NewFactory(
	&Host{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.Alphanumeric(randomdata.Number(5, 50)), nil
}).OnCreate(func(args factory.Args) error {
	host := args.Instance().(*Host)
	host.Interfaces[0].BMC = false
	host.Interfaces[1].BMC = true
	return nil
}).SubSliceFactory("Interfaces", NetInterfaceFactory, func() int { return 2 })

func TestFactory(t *testing.T) {
	assert := assert.New(t)

	for i := 0; i < 3; i++ {
		host := HostFactory.MustCreate().(*Host)
		assert.Greater(len(host.Name), 1)
		assert.Equal(2, len(host.Interfaces))
	}
}
