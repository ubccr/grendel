// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gosnmp/gosnmp"
	"github.com/ubccr/grendel/pkg/model"
)

const (
	oidDot1qTpFdbPort        = ".1.3.6.1.2.1.17.7.1.2.2.1.2"
	oidDot1qTpFdbStatus      = ".1.3.6.1.2.1.17.7.1.2.2.1.3"
	oidDot1qVlanFdbID        = ".1.3.6.1.2.1.17.7.1.4.2.1.3"
	oidDot1dBasePortIfIndex  = ".1.3.6.1.2.1.17.1.4.1.2"
	oidIfName                = ".1.3.6.1.2.1.31.1.1.1.1"
	oidIfAlias               = ".1.3.6.1.2.1.31.1.1.1.18"
	oidLldpLocPortID         = ".1.0.8802.1.1.2.1.3.7.1.3"
	oidLldpLocPortDesc       = ".1.0.8802.1.1.2.1.3.7.1.4"
	oidLldpRemTable          = ".1.0.8802.1.1.2.1.4.1.1"
	oidLldpRemManAddrIfIndex = ".1.0.8802.1.1.2.1.4.2.1.4"
)

// dot1qTpFdbStatus values from the Q-BRIDGE MIB; self(4) is the switch's own address.
const (
	fdbStatusSelf = 4
	fdbStatusMgmt = 5
)

var lldpChassisSubtype = map[int]string{1: "CHASSIS_COMPONENT", 2: "INTERFACE_ALIAS", 3: "PORT_COMPONENT", 4: "MAC_ADDRESS", 5: "NETWORK_ADDRESS", 6: "INTERFACE_NAME", 7: "LOCAL"}

var lldpPortSubtype = map[int]string{1: "INTERFACE_ALIAS", 2: "PORT_COMPONENT", 3: "MAC_ADDRESS", 4: "NETWORK_ADDRESS", 5: "INTERFACE_NAME", 6: "AGENT_CIRCUIT_ID", 7: "LOCAL"}

// SNMP keeps one session and every table it has walked, so the calls of one scan share a connection and walk ifName once
type SNMP struct {
	addr string
	cfg  SNMPConfig

	mu     sync.Mutex
	conn   *gosnmp.GoSNMP
	tables map[string][]gosnmp.SnmpPDU
}

func NewSNMP(addr string, cfg Config) (*SNMP, error) {
	switch cfg.SNMP.Version {
	case "2c", "3":
	default:
		return nil, fmt.Errorf("unsupported switch.snmp.version %q, use 2c or 3", cfg.SNMP.Version)
	}
	return &SNMP{addr: addr, cfg: cfg.SNMP, tables: map[string][]gosnmp.SnmpPDU{}}, nil
}

func (s *SNMP) connect(ctx context.Context) (*gosnmp.GoSNMP, error) {
	c := &gosnmp.GoSNMP{
		Context:            ctx,
		Target:             s.addr,
		Port:               uint16(s.cfg.Port),
		Timeout:            s.cfg.Timeout,
		Retries:            s.cfg.Retries,
		MaxOids:            gosnmp.MaxOids,
		MaxRepetitions:     25,
		ExponentialTimeout: true,
	}

	if s.cfg.Version == "3" {
		usm := &gosnmp.UsmSecurityParameters{
			UserName:                 s.cfg.User,
			AuthenticationPassphrase: s.cfg.AuthPass,
			PrivacyPassphrase:        s.cfg.PrivPass,
		}
		flags := gosnmp.NoAuthNoPriv
		if s.cfg.AuthPass != "" {
			p, err := authProto(s.cfg.AuthProto)
			if err != nil {
				return nil, err
			}
			usm.AuthenticationProtocol = p
			flags = gosnmp.AuthNoPriv
		}
		if s.cfg.PrivPass != "" {
			p, err := privProto(s.cfg.PrivProto)
			if err != nil {
				return nil, err
			}
			usm.PrivacyProtocol = p
			flags = gosnmp.AuthPriv
		}
		c.Version = gosnmp.Version3
		c.SecurityModel = gosnmp.UserSecurityModel
		c.MsgFlags = flags
		c.SecurityParameters = usm
	} else {
		c.Version = gosnmp.Version2c
		c.Community = s.cfg.Community
	}

	if err := c.Connect(); err != nil {
		return nil, fmt.Errorf("snmp connect to %s: %w", s.addr, err)
	}
	return c, nil
}

func authProto(p string) (gosnmp.SnmpV3AuthProtocol, error) {
	switch strings.ToUpper(p) {
	case "MD5":
		return gosnmp.MD5, nil
	case "SHA", "":
		return gosnmp.SHA, nil
	case "SHA224":
		return gosnmp.SHA224, nil
	case "SHA256":
		return gosnmp.SHA256, nil
	case "SHA384":
		return gosnmp.SHA384, nil
	case "SHA512":
		return gosnmp.SHA512, nil
	}
	return 0, fmt.Errorf("unsupported switch.snmp.auth_proto %q", p)
}

func privProto(p string) (gosnmp.SnmpV3PrivProtocol, error) {
	switch strings.ToUpper(p) {
	case "DES":
		return gosnmp.DES, nil
	case "AES", "":
		return gosnmp.AES, nil
	case "AES192":
		return gosnmp.AES192, nil
	case "AES256":
		return gosnmp.AES256, nil
	case "AES192C":
		return gosnmp.AES192C, nil
	case "AES256C":
		return gosnmp.AES256C, nil
	}
	return 0, fmt.Errorf("unsupported switch.snmp.priv_proto %q", p)
}

// walk bulk-walks each table and keys the results by the table OID, so callers can tell which column a PDU came from. A gosnmp session isn't safe for concurrent use, so walks take turns on it.
func (s *SNMP) walk(ctx context.Context, oids ...string) (map[string][]gosnmp.SnmpPDU, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		c, err := s.connect(ctx)
		if err != nil {
			return nil, err
		}
		s.conn = c
	}
	s.conn.Context = ctx

	out := map[string][]gosnmp.SnmpPDU{}
	for _, oid := range oids {
		pdus, ok := s.tables[oid]
		if !ok {
			var err error
			if pdus, err = s.conn.BulkWalkAll(oid); err != nil {
				return nil, fmt.Errorf("snmp walk %s on %s: %w", oid, s.addr, err)
			}
			s.tables[oid] = pdus
		}
		out[oid] = pdus
	}
	return out, nil
}

func (s *SNMP) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return nil
	}
	err := s.conn.Conn.Close()
	s.conn = nil
	return err
}

func (s *SNMP) GetInterfaceDescriptions(ctx context.Context) (map[string]string, error) {
	w, err := s.walk(ctx, oidIfName, oidIfAlias)
	if err != nil {
		return nil, err
	}
	return parseIfTable(w), nil
}

func (s *SNMP) GetMACTable(ctx context.Context) (model.MACTable, error) {
	w, err := s.walk(ctx, oidDot1qTpFdbPort, oidDot1qTpFdbStatus, oidDot1qVlanFdbID, oidDot1dBasePortIfIndex, oidIfName)
	if err != nil {
		return nil, err
	}
	return parseSNMPFDB(w), nil
}

func (s *SNMP) GetLLDPNeighbors(ctx context.Context) (model.LLDPNeighbors, error) {
	w, err := s.walk(ctx, oidLldpRemTable, oidLldpRemManAddrIfIndex, oidLldpLocPortID, oidLldpLocPortDesc, oidIfName)
	if err != nil {
		return nil, err
	}
	return parseSNMPLLDP(w), nil
}

// index returns the OID suffix after table as integers, or nil when the PDU is outside the table.
func index(pdu gosnmp.SnmpPDU, table string) []int {
	rest, ok := strings.CutPrefix(pdu.Name, table+".")
	if !ok {
		return nil
	}
	parts := strings.Split(rest, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		out = append(out, n)
	}
	return out
}

func pduInt(pdu gosnmp.SnmpPDU) int {
	v := gosnmp.ToBigInt(pdu.Value)
	if v == nil || !v.IsInt64() {
		return 0
	}
	return int(v.Int64())
}

func pduBytes(pdu gosnmp.SnmpPDU) []byte {
	switch v := pdu.Value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	}
	return nil
}

// pduText decodes an octet string that may hold text or binary, such as an LLDP port ID.
func pduText(b []byte) string {
	s := strings.TrimRight(string(b), "\x00")
	if utf8.ValidString(s) && !strings.ContainsFunc(s, func(r rune) bool { return r < 0x20 }) {
		return s
	}
	return net.HardwareAddr(b).String()
}

func ifNames(w map[string][]gosnmp.SnmpPDU) map[int]string {
	out := map[int]string{}
	for _, pdu := range w[oidIfName] {
		if idx := index(pdu, oidIfName); len(idx) == 1 {
			out[idx[0]] = string(pduBytes(pdu))
		}
	}
	return out
}

func parseIfTable(w map[string][]gosnmp.SnmpPDU) map[string]string {
	names := ifNames(w)
	out := map[string]string{}
	for _, name := range names {
		out[name] = ""
	}
	for _, pdu := range w[oidIfAlias] {
		if idx := index(pdu, oidIfAlias); len(idx) == 1 {
			if name, ok := names[idx[0]]; ok {
				out[name] = string(pduBytes(pdu))
			}
		}
	}
	return out
}

func macFromIndex(idx []int) (net.HardwareAddr, bool) {
	if len(idx) != 6 {
		return nil, false
	}
	mac := make(net.HardwareAddr, 6)
	for i, b := range idx {
		if b < 0 || b > 255 {
			return nil, false
		}
		mac[i] = byte(b)
	}
	return mac, true
}

func parseSNMPFDB(w map[string][]gosnmp.SnmpPDU) model.MACTable {
	names := ifNames(w)

	bridgeToIf := map[int]int{}
	for _, pdu := range w[oidDot1dBasePortIfIndex] {
		if idx := index(pdu, oidDot1dBasePortIfIndex); len(idx) == 1 {
			bridgeToIf[idx[0]] = pduInt(pdu)
		}
	}

	// dot1qVlanFdbId is indexed by (timeMark, vlan) and holds the FDB id; most switches use the VLAN id as the FDB id, so that is the fallback.
	fdbToVlan := map[int]int{}
	for _, pdu := range w[oidDot1qVlanFdbID] {
		if idx := index(pdu, oidDot1qVlanFdbID); len(idx) == 2 {
			fdbToVlan[pduInt(pdu)] = idx[1]
		}
	}

	status := map[string]int{}
	for _, pdu := range w[oidDot1qTpFdbStatus] {
		status[strings.TrimPrefix(pdu.Name, oidDot1qTpFdbStatus)] = pduInt(pdu)
	}

	out := model.MACTable{}
	for _, pdu := range w[oidDot1qTpFdbPort] {
		idx := index(pdu, oidDot1qTpFdbPort)
		if len(idx) != 7 {
			continue
		}
		mac, ok := macFromIndex(idx[1:])
		if !ok {
			continue
		}
		bridgePort := pduInt(pdu)
		st := status[strings.TrimPrefix(pdu.Name, oidDot1qTpFdbPort)]
		if bridgePort == 0 || st == fdbStatusSelf {
			continue
		}

		ifIndex, ok := bridgeToIf[bridgePort]
		// only a switch with no dot1dBasePortIfIndex table numbers its bridge ports by ifIndex; in a table that has rows, a port missing from it is not an interface
		if !ok && len(bridgeToIf) == 0 {
			ifIndex, ok = bridgePort, true
		}
		name, named := names[ifIndex]
		if !ok || !named {
			log.Debugf("fdb entry %s on bridge port %d has no interface name", mac, bridgePort)
			continue
		}

		vlan, ok := fdbToVlan[idx[0]]
		if !ok {
			vlan = idx[0]
		}

		typ := "dynamic"
		if st == fdbStatusMgmt {
			typ = "static"
		}

		out = append(out, &model.MACTableEntry{Ifname: name, VLAN: strconv.Itoa(vlan), Type: typ, MAC: mac})
	}
	return out
}

func parseSNMPLLDP(w map[string][]gosnmp.SnmpPDU) model.LLDPNeighbors {
	names := ifNames(w)

	// lldpLocPortNum is usually the ifIndex, but the MIB does not require it; lldpLocPortId or its description names the port when it is not.
	locPort := map[int]string{}
	for _, pdu := range w[oidLldpLocPortDesc] {
		if idx := index(pdu, oidLldpLocPortDesc); len(idx) == 1 {
			locPort[idx[0]] = string(pduBytes(pdu))
		}
	}
	for _, pdu := range w[oidLldpLocPortID] {
		if idx := index(pdu, oidLldpLocPortID); len(idx) == 1 {
			if id := pduText(pduBytes(pdu)); id != "" && !strings.Contains(id, ":") {
				locPort[idx[0]] = id
			}
		}
	}
	portName := func(n int) string {
		if name, ok := names[n]; ok {
			return name
		}
		if name, ok := locPort[n]; ok && name != "" {
			return name
		}
		return strconv.Itoa(n)
	}

	type remKey struct{ local, index int }
	rows := map[remKey]map[int]gosnmp.SnmpPDU{}
	for _, pdu := range w[oidLldpRemTable] {
		// lldpRemEntry.<column>.<timeMark>.<localPortNum>.<index>
		idx := index(pdu, oidLldpRemTable)
		if len(idx) != 4 {
			continue
		}
		k := remKey{idx[2], idx[3]}
		if rows[k] == nil {
			rows[k] = map[int]gosnmp.SnmpPDU{}
		}
		rows[k][idx[0]] = pdu
	}

	mgmt := map[remKey]string{}
	for _, pdu := range w[oidLldpRemManAddrIfIndex] {
		// lldpRemManAddrEntry.<column>.<timeMark>.<localPortNum>.<index>.<addrSubtype>.<addrLen>.<addr...>
		idx := index(pdu, oidLldpRemManAddrIfIndex)
		if len(idx) != 9 || idx[3] != 1 || idx[4] != 4 {
			continue
		}
		mgmt[remKey{idx[1], idx[2]}] = fmt.Sprintf("%d.%d.%d.%d", idx[5], idx[6], idx[7], idx[8])
	}

	// a port can have several neighbors, such as a shared-LOM host and its BMC, but the result holds one per port: the lowest remote index, so every call picks the same one
	keys := slices.SortedFunc(maps.Keys(rows), func(a, b remKey) int {
		return cmp.Or(cmp.Compare(a.local, b.local), cmp.Compare(a.index, b.index))
	})

	out := model.LLDPNeighbors{}
	for _, k := range keys {
		cols := rows[k]
		str := func(c int) string { return pduText(pduBytes(cols[c])) }
		chassisType := lldpChassisSubtype[pduInt(cols[4])]
		portType := lldpPortSubtype[pduInt(cols[6])]

		chassis := str(5)
		if chassisType == "MAC_ADDRESS" {
			chassis = net.HardwareAddr(pduBytes(cols[5])).String()
		}
		portID := str(7)
		if portType == "MAC_ADDRESS" {
			portID = net.HardwareAddr(pduBytes(cols[7])).String()
		}

		name := portName(k.local)
		if _, seen := out[name]; seen {
			continue
		}
		out[name] = &model.LLDP{
			PortName:          name,
			ChassisIdType:     chassisType,
			ChassisId:         chassis,
			PortIdType:        portType,
			PortId:            portID,
			PortDescription:   string(pduBytes(cols[8])),
			SystemName:        string(pduBytes(cols[9])),
			SystemDescription: string(pduBytes(cols[10])),
			ManagementAddress: mgmt[k],
		}
	}
	return out
}
