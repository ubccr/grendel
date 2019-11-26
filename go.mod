module github.com/ubccr/grendel

go 1.13

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c // indirect
	github.com/hugelgupf/socketpair v0.0.0-20190730060125-05d35a94e714 // indirect
	github.com/insomniacslk/dhcp v0.0.0-20191120203056-ec0e0154d15c
	github.com/labstack/echo/v4 v4.1.11
	github.com/mdlayher/ethernet v0.0.0-20190606142754-0394541c37b7 // indirect
	github.com/mdlayher/raw v0.0.0-20190606144222-a54781e5f38f // indirect
	github.com/pin/tftp v2.1.0+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/square/certstrap v1.2.0
	github.com/u-root/u-root v5.0.0+incompatible // indirect
	github.com/urfave/cli v1.22.1
	go.universe.tf/netboot v0.0.0-20190802213723-72fa512fed0f
	golang.org/x/crypto v0.0.0-20191117063200-497ca9f6d64f // indirect
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65
	golang.org/x/sys v0.0.0-20191119060738-e882bf8e40c2 // indirect
	gopkg.in/yaml.v2 v2.2.4 // indirect
)

replace github.com/square/certstrap => github.com/ubccr/certstrap v1.2.1-0.20191119164315-e73497b6d54c
