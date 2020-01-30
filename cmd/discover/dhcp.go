package discover

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/nodeset"
)

type discoveryDHCP struct {
	nodeset *nodeset.NodeSetIterator
	seen    map[string]bool
	subnet  net.IP
	netmask net.IPMask
}

var (
	trace            bool
	nodeNumberRegexp = regexp.MustCompile(`(\d+)$`)
	dhcpCmd          = &cobra.Command{
		Use:   "dhcp",
		Short: "Auto-discover hosts from DHCP",
		Long:  `Auto-discover hosts from DHCP`,
		RunE: func(command *cobra.Command, args []string) error {
			if trace {
				log.Infof("Tracing DHCP packets on %s", viper.GetString("discovery.listen"))
				snooper, err := dhcp.NewSnooper(viper.GetString("discovery.listen"), traceDHCP)
				if err != nil {
					return err
				}

				return snooper.Snoop()
			}

			if subnetStr == "" {
				return fmt.Errorf("Please provide a subnet (--subnet)")
			}

			netmask := net.IPv4Mask(255, 255, 255, 0)
			subnet := net.ParseIP(subnetStr)
			if subnet == nil || subnet.To4() == nil {
				return fmt.Errorf("Invalid IPv4 subnet address: %s", subnetStr)
			}

			if len(args) == 0 {
				return fmt.Errorf("Please provide a nodeset")
			}

			ns, err := nodeset.NewNodeSet(strings.Join(args, ","))
			if err != nil {
				return err
			}

			d := &discoveryDHCP{
				nodeset: ns.Iterator(),
				seen:    make(map[string]bool),
				subnet:  subnet,
				netmask: netmask,
			}

			snooper, err := dhcp.NewSnooper(viper.GetString("discovery.listen"), d.handler)
			if err != nil {
				return err
			}

			return snooper.Snoop()
		},
	}
)

func init() {
	dhcpCmd.Flags().StringP("listen", "l", "0.0.0.0:67", "address to run discovery DHCP server")
	viper.BindPFlag("discovery.listen", dhcpCmd.Flags().Lookup("listen"))

	dhcpCmd.Flags().BoolVar(&trace, "trace", false, "Print DHCP packets only")

	discoverCmd.AddCommand(dhcpCmd)
}

func traceDHCP(req *dhcpv4.DHCPv4) {
	log.Debugf("Received DHCPv4 packet")
	log.Debugf(req.Summary())
}

func (d *discoveryDHCP) handler(req *dhcpv4.DHCPv4) {
	log.Debugf("Received DHCPv4 packet")
	log.Debugf(req.Summary())

	if req.OpCode != dhcpv4.OpcodeBootRequest {
		log.Warningf("not a BootRequest, ignoring")
		return
	}

	if req.MessageType() != dhcpv4.MessageTypeDiscover {
		log.Warnf("Discovery unhandled message type: %v", req.MessageType())
		return
	}

	if _, ok := d.seen[req.ClientHWAddr.String()]; ok {
		log.Infof("Already seen mac address. skipping: %s", req.ClientHWAddr)
		return
	}

	if !d.nodeset.Next() {
		log.Errorf("No more values in nodeset")
		return
	}

	d.seen[req.ClientHWAddr.String()] = true

	ip := d.subnet.Mask(d.netmask)
	matches := nodeNumberRegexp.FindStringSubmatch(d.nodeset.Value())
	if len(matches) != 2 {
		log.Errorf("node doesn't end in number. failed to generate IP address: %s", d.nodeset.Value())
		return
	}
	num, _ := strconv.Atoi(matches[1])
	ip[3] += uint8(num)

	fmt.Printf("%s\t%s\t%s\n", d.nodeset.Value(), req.ClientHWAddr, ip.String())
}
