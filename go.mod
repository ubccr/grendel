module github.com/ubccr/grendel

go 1.13

require (
	github.com/GeertJohan/go.rice v1.0.0
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-playground/validator/v10 v10.0.1
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c // indirect
	github.com/hugelgupf/socketpair v0.0.0-20190730060125-05d35a94e714 // indirect
	github.com/insomniacslk/dhcp v0.0.0-20191120203056-ec0e0154d15c
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/labstack/echo/v4 v4.1.11
	github.com/labstack/gommon v0.3.0
	github.com/mdlayher/ethernet v0.0.0-20190606142754-0394541c37b7 // indirect
	github.com/mdlayher/raw v0.0.0-20190606144222-a54781e5f38f // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/miekg/dns v1.1.24
	github.com/onsi/ginkgo v1.10.3 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/pin/tftp v2.1.0+incompatible
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/viper v1.3.2
	github.com/square/certstrap v1.2.0
	github.com/stmcginnis/gofish v0.2.0
	github.com/timshannon/badgerhold v0.0.0-20191028155359-570a6f824389
	github.com/u-root/u-root v5.0.0+incompatible // indirect
	github.com/urfave/cli v1.22.1
	github.com/vmware/goipmi v0.0.0-20181114221114-2333cd82d702
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/crypto v0.0.0-20191117063200-497ca9f6d64f // indirect
	golang.org/x/net v0.0.0-20191003171128-d98b1b443823 // indirect
	golang.org/x/sys v0.0.0-20191119060738-e882bf8e40c2 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/square/certstrap => github.com/ubccr/certstrap v1.2.1-0.20191119164315-e73497b6d54c
