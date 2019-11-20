package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/firmware"
	"github.com/urfave/cli"
	"go.universe.tf/netboot/pixiecore"
)

func NewServeCommand() cli.Command {
	return cli.Command{
		Name:        "serve",
		Usage:       "Start PXE server",
		Description: "Start PXE server",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "kernel",
				Usage: "Location of kernel vmlinuz file",
			},
			cli.StringSliceFlag{
				Name:  "initrd",
				Usage: "Location of initrd file(s)",
			},
			cli.StringFlag{
				Name:  "cmdline",
				Usage: "Kernel commandline arguments",
			},
			cli.StringFlag{
				Name:  "bootmsg",
				Usage: "Message to print on machines before booting",
			},
			cli.StringFlag{
				Name:  "listen-address",
				Value: "0.0.0.0",
				Usage: "IPv4 address to listen on",
			},
			cli.StringFlag{
				Name:  "cert",
				Usage: "Path to certificate",
			},
			cli.StringFlag{
				Name:  "key",
				Usage: "Path to private key",
			},
		},
		Action: runServices,
	}
}

func logger(subsystem, msg string) {
	logrus.WithFields(logrus.Fields{
		"subsystem": subsystem,
	}).Info(msg)
}

func debugger(subsystem, msg string) {
	logrus.WithFields(logrus.Fields{
		"subsystem": subsystem,
	}).Debug(msg)
}

func runServices(c *cli.Context) error {
	spec := &pixiecore.Spec{
		Kernel:  pixiecore.ID(c.String("kernel")),
		Cmdline: c.String("cmdline"),
		Message: c.String("bootmsg"),
	}

	for _, initrd := range c.StringSlice("initrd") {
		spec.Initrd = append(spec.Initrd, pixiecore.ID(initrd))
	}

	booter, err := pixiecore.StaticBooter(spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't make static booter: %s\n", err)
		os.Exit(1)
	}

	server := &api.Server{
		Ipxe:     map[pixiecore.Firmware][]byte{},
		Log:      logger,
		Debug:    debugger,
		Booter:   booter,
		Address:  c.String("listen-address"),
		CertFile: c.String("cert"),
		KeyFile:  c.String("key"),
	}

	if server.CertFile != "" && server.KeyFile != "" {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch hostname: %s\n", err)
			os.Exit(1)
		}
		server.Hostname = hostname
	}

	server.Ipxe[pixiecore.FirmwareX86PC] = firmware.MustAsset("undionly.kpxe")
	server.Ipxe[pixiecore.FirmwareEFI32] = firmware.MustAsset("ipxe-i386.efi")
	server.Ipxe[pixiecore.FirmwareEFI64] = firmware.MustAsset("snponly-x86_64.efi")
	server.Ipxe[pixiecore.FirmwareEFIBC] = firmware.MustAsset("snponly-x86_64.efi")
	server.Ipxe[pixiecore.FirmwareX86Ipxe] = firmware.MustAsset("ipxe.pxe")

	return server.Serve()
}

func runPXEServer(c *cli.Context) error {
	spec := &pixiecore.Spec{
		Kernel:  pixiecore.ID(c.String("kernel")),
		Cmdline: c.String("cmdline"),
		Message: c.String("bootmsg"),
	}

	for _, initrd := range c.StringSlice("initrd") {
		spec.Initrd = append(spec.Initrd, pixiecore.ID(initrd))
	}

	booter, err := pixiecore.StaticBooter(spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't make static booter: %s\n", err)
		os.Exit(1)
	}

	server := &pixiecore.Server{
		Ipxe:    map[pixiecore.Firmware][]byte{},
		Log:     logger,
		Debug:   debugger,
		Booter:  booter,
		Address: c.String("listen-address"),
	}

	server.Ipxe[pixiecore.FirmwareX86PC] = firmware.MustAsset("undionly.kpxe")
	server.Ipxe[pixiecore.FirmwareEFI32] = firmware.MustAsset("ipxe-i386.efi")
	server.Ipxe[pixiecore.FirmwareEFI64] = firmware.MustAsset("snponly-x86_64.efi")
	server.Ipxe[pixiecore.FirmwareEFIBC] = firmware.MustAsset("snponly-x86_64.efi")
	server.Ipxe[pixiecore.FirmwareX86Ipxe] = firmware.MustAsset("ipxe.pxe")

	return server.Serve()
}
