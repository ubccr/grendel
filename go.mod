module github.com/ubccr/grendel

go 1.13

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c // indirect
	github.com/labstack/echo/v4 v4.1.11
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.5.0
	github.com/square/certstrap v1.2.0
	github.com/urfave/cli v1.22.1
	go.universe.tf/netboot v0.0.0-20190802213723-72fa512fed0f
	golang.org/x/crypto v0.0.0-20191117063200-497ca9f6d64f // indirect
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092
	golang.org/x/sys v0.0.0-20191119060738-e882bf8e40c2 // indirect
)

replace github.com/square/certstrap => github.com/ubccr/certstrap v1.2.1-0.20191119164315-e73497b6d54c
