module github.com/ubccr/grendel

go 1.16

require (
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/alouca/gologger v0.0.0-20120904114645-7d4b7291de9c // indirect
	github.com/alouca/gosnmp v0.0.0-20170620005048-04d83944c9ab
	github.com/aws/aws-sdk-go v1.44.19 // indirect
	github.com/bits-and-blooms/bitset v1.2.1
	github.com/bluele/factory-go v0.0.0-20181130035244-e6e8633dd3fe
	github.com/clarketm/json v1.17.1 // indirect
	github.com/coreos/butane v0.14.1-0.20220513204719-6cd92788076e
	github.com/coreos/go-json v0.0.0-20220325222439-31b2177291ae // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/coreos/vcontext v0.0.0-20220326205524-7fcaf69e7050 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.10.0
	github.com/go-playground/validator/v10 v10.0.1
	github.com/hako/branca v0.0.0-20191227164554-3b9970524189
	github.com/hashicorp/go-retryablehttp v0.6.6
	github.com/insomniacslk/dhcp v0.0.0-20200922210017-67c425063dca
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/korovkin/limiter v0.0.0-20190919045942-dac5a6b2a536
	github.com/labstack/echo/v4 v4.1.11
	github.com/labstack/gommon v0.3.0
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/miekg/dns v1.1.43
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pin/tftp v2.1.1-0.20200117065540-2f79be2dba4e+incompatible
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/segmentio/fasthash v1.0.3
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.6.2
	github.com/stmcginnis/gofish v0.2.0
	github.com/stretchr/testify v1.7.1
	github.com/tidwall/buntdb v1.1.2
	github.com/tidwall/gjson v1.10.2
	github.com/tidwall/sjson v1.1.7
	github.com/ubccr/go-dhcpd-leases v0.1.1-0.20191206204522-601ab01835fb
	github.com/vmware/goipmi v0.0.0-20181114221114-2333cd82d702
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/yaml.v3 v3.0.0 // indirect
)

replace github.com/square/certstrap => github.com/ubccr/certstrap v1.2.1-0.20200115142812-0c31c5e59383

replace github.com/pin/tftp => github.com/ubccr/tftp v0.0.0-20200215220641-2b5a116e0866
