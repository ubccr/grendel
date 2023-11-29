package frontend

import (
	"fmt"
	"net/netip"

	"github.com/ubccr/grendel/model"
	"go4.org/netipx"
)

func (h *Handler) newHostIPs(subnet string) ([]string, error) {
	var ipRange []string

	hostList, err := h.DB.Hosts()
	if err != nil {
		return ipRange, err
	}

	var builder netipx.IPSetBuilder
	ipmap := make(map[netip.Addr]*model.Host)
	ipp, err := netip.ParsePrefix(subnet)

	// Set aside first and last IP for
	firstIp := ipp.Addr()
	lastIp := netipx.PrefixLastIP(ipp)

	if err != nil {
		return ipRange, err
	}

	builder.AddPrefix(ipp)
	iset, _ := builder.IPSet()

	for _, host := range hostList {
		for _, i := range host.Interfaces {
			if iset.Contains(i.IP.Addr()) {
				builder.Remove(i.IP.Addr())
				ipmap[i.IP.Addr()] = host
			}
		}
	}

	iset, _ = builder.IPSet()
	prefixes := iset.Prefixes()

	for _, p := range prefixes {
		i := p.Addr()
		last := netipx.PrefixLastIP(p)

		for ; i.Compare(last) <= 0; i = i.Next() {
			if i.Compare(firstIp) != 0 && i.Compare(lastIp) != 0 {
				// TODO: set from config or pull from form
				ipRange = append(ipRange, fmt.Sprintf("%s/%s", i.String(), "20"))
			} else {
				continue
			}
		}

	}

	return ipRange, nil
}
