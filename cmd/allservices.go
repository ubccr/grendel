package cmd

import (
	"os"
	"reflect"

	"github.com/ubccr/grendel/model"
	"github.com/urfave/cli/v2"
)

func flagExists(flags []cli.Flag, flag cli.Flag) bool {
	for _, f := range flags {
		if reflect.DeepEqual(flag.Names(), f.Names()) {
			return true
		}
	}

	return false
}

func NewServeAllCommand() *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "kernel",
			Usage: "Location of kernel vmlinuz file",
		},
		&cli.StringSliceFlag{
			Name:  "initrd",
			Usage: "Location of initrd file(s)",
		},
		&cli.StringFlag{
			Name:  "cmdline",
			Usage: "Kernel commandline arguments",
		},
		&cli.StringFlag{
			Name:  "liveimg",
			Usage: "Location of liveimg squashfs",
		},
		&cli.StringFlag{
			Name:  "rootfs",
			Usage: "Location of rootfs",
		},
		&cli.StringFlag{
			Name:  "install-repo",
			Usage: "URL of repo mirror",
		},
		&cli.StringFlag{
			Name:  "static-hosts",
			Usage: "static hosts file",
		},
		&cli.StringFlag{
			Name:  "dhcp-leases",
			Usage: "dhcp leases file",
		},
		&cli.StringFlag{
			Name:  "json-hosts",
			Usage: "json hosts file",
		},
	}

	for _, cmd := range []*cli.Command{NewDHCPCommand(), NewTFTPCommand(), NewAPICommand(), NewProvisionCommand(), NewDNSCommand()} {
		for _, f := range cmd.Flags {
			if !flagExists(flags, f) {
				flags = append(flags, f)
			}
		}
	}

	return &cli.Command{
		Name:        "all",
		Usage:       "Start all services",
		Description: "Start all services",
		Flags:       flags,
		Action:      runAllServices,
	}
}

func runAllServices(c *cli.Context) error {
	staticBooter, err := model.NewStaticBooter(c.String("kernel"), c.StringSlice("initrd"), c.String("cmdline"), c.String("liveimg"), c.String("rootfs"), c.String("install-repo"))
	if err != nil {
		return err
	}

	if c.IsSet("static-hosts") {
		file, err := os.Open(c.String("static-hosts"))
		if err != nil {
			return err
		}
		defer file.Close()

		err = staticBooter.LoadStaticHosts(file)
		if err != nil {
			return err
		}
	}

	if c.IsSet("dhcp-leases") {
		file, err := os.Open(c.String("dhcp-leases"))
		if err != nil {
			return err
		}
		defer file.Close()

		err = staticBooter.LoadDHCPLeases(file)
		if err != nil {
			return err
		}
	}

	if c.IsSet("json-hosts") {
		file, err := os.Open(c.String("json-hosts"))
		if err != nil {
			return err
		}
		defer file.Close()

		err = staticBooter.LoadJSON(file)
		if err != nil {
			return err
		}
	}

	if !c.IsSet("socket-path") {
		c.Set("socket-path", "grendel-api.socket")
	}

	DB = staticBooter

	errs := make(chan error, 5)

	go func() { errs <- runAPI(c) }()
	go func() { errs <- runDHCP(c) }()
	go func() { errs <- runTFTP(c) }()
	go func() { errs <- runProvision(c) }()
	go func() { errs <- runDNS(c) }()

	err = <-errs
	return err
}
