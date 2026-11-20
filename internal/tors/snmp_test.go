// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"fmt"
	"testing"

	"github.com/gosnmp/gosnmp"
)

func intPDU(oid string, v int) gosnmp.SnmpPDU {
	return gosnmp.SnmpPDU{Name: oid, Type: gosnmp.Integer, Value: v}
}

func strPDU(oid string, v string) gosnmp.SnmpPDU {
	return gosnmp.SnmpPDU{Name: oid, Type: gosnmp.OctetString, Value: []byte(v)}
}

func bytesPDU(oid string, v []byte) gosnmp.SnmpPDU {
	return gosnmp.SnmpPDU{Name: oid, Type: gosnmp.OctetString, Value: v}
}

// fdbOID builds dot1qTpFdb<column>.<fdbId>.<mac as six decimals>.
func fdbOID(table string, fdb int, mac [6]int) string {
	return fmt.Sprintf("%s.%d.%d.%d.%d.%d.%d.%d", table, fdb, mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

func os10Walk() map[string][]gosnmp.SnmpPDU {
	bmc1 := [6]int{0xd0, 0x8e, 0x79, 0, 0, 1}
	bmc4 := [6]int{0xd0, 0x8e, 0x79, 0, 0, 4}
	lom4 := [6]int{0xb8, 0xca, 0x3a, 0, 0, 4}
	self := [6]int{0x74, 0x83, 0xef, 0, 0, 0}
	return map[string][]gosnmp.SnmpPDU{
		oidIfName: {
			strPDU(oidIfName+".1001", "ethernet1/1/1"),
			strPDU(oidIfName+".1004", "ethernet1/1/4"),
			strPDU(oidIfName+".9000", "vlan3001"),
		},
		oidDot1dBasePortIfIndex: {
			intPDU(oidDot1dBasePortIfIndex+".1", 1001),
			intPDU(oidDot1dBasePortIfIndex+".4", 1004),
		},
		// fdb id 7 is vlan 3001 on this switch, so the fdb id cannot be read as the vlan.
		oidDot1qVlanFdbID: {
			intPDU(oidDot1qVlanFdbID+".0.3001", 7),
		},
		oidDot1qTpFdbPort: {
			intPDU(fdbOID(oidDot1qTpFdbPort, 7, bmc1), 1),
			intPDU(fdbOID(oidDot1qTpFdbPort, 7, bmc4), 4),
			intPDU(fdbOID(oidDot1qTpFdbPort, 7, lom4), 4),
			intPDU(fdbOID(oidDot1qTpFdbPort, 7, self), 0),
			intPDU(fdbOID(oidDot1qTpFdbPort, 12, bmc1), 99),
		},
		oidDot1qTpFdbStatus: {
			intPDU(fdbOID(oidDot1qTpFdbStatus, 7, bmc1), 3),
			intPDU(fdbOID(oidDot1qTpFdbStatus, 7, bmc4), 3),
			intPDU(fdbOID(oidDot1qTpFdbStatus, 7, lom4), 5),
			intPDU(fdbOID(oidDot1qTpFdbStatus, 7, self), 4),
		},
	}
}

func TestSNMPFDB(t *testing.T) {
	table := parseSNMPFDB(os10Walk())
	by := table.ByIfname()

	if len(by["ethernet1/1/1"]) != 1 || by["ethernet1/1/1"][0].MAC.String() != "d0:8e:79:00:00:01" {
		t.Errorf("ethernet1/1/1 = %v", by["ethernet1/1/1"])
	}
	if by["ethernet1/1/1"][0].VLAN != "3001" {
		t.Errorf("vlan = %q, want the fdb id mapped to 3001", by["ethernet1/1/1"][0].VLAN)
	}
	if len(by["ethernet1/1/4"]) != 2 {
		t.Errorf("ethernet1/1/4 should carry both shared-LOM macs: %v", by["ethernet1/1/4"])
	}
	for _, e := range by["ethernet1/1/4"] {
		if e.MAC.String() == "b8:ca:3a:00:00:04" && e.Type != "static" {
			t.Errorf("mgmt status should read as static, got %q", e.Type)
		}
	}
	if len(table) != 3 {
		t.Errorf("got %d entries, want 3: the self entry and the one on an unknown bridge port are dropped", len(table))
	}
}

func TestSNMPFDBFallsBackToFdbIDAndBridgePort(t *testing.T) {
	w := os10Walk()
	delete(w, oidDot1qVlanFdbID)
	delete(w, oidDot1dBasePortIfIndex)
	w[oidIfName] = append(w[oidIfName], strPDU(oidIfName+".1", "ethernet1/1/1"))

	by := parseSNMPFDB(w).ByIfname()
	if e := by["ethernet1/1/1"]; len(e) != 1 || e[0].VLAN != "7" {
		t.Errorf("without the vlan and bridge maps, fdb id and bridge port are used as is: %v", e)
	}
}

func TestSNMPInterfaces(t *testing.T) {
	w := map[string][]gosnmp.SnmpPDU{
		oidIfName:        {strPDU(oidIfName+".1001", "ethernet1/1/1")},
		oidIfOperStatus:  {intPDU(oidIfOperStatus+".1001", 1)},
		oidIfAdminStatus: {intPDU(oidIfAdminStatus+".1001", 1)},
		oidIfHighSpeed:   {gosnmp.SnmpPDU{Name: oidIfHighSpeed + ".1001", Type: gosnmp.Gauge32, Value: uint(1000)}},
		oidIfMtu:         {intPDU(oidIfMtu+".1001", 9216)},
		oidIfAlias:       {strPDU(oidIfAlias+".1001", "cpn-a01-01")},
	}
	st := parseIfTable(w)["ethernet1/1/1"]
	if st == nil || st.LineProtocolStatus != "up" || st.Bandwidth != 1000 || st.MTU != 9216 || st.Description != "cpn-a01-01" {
		t.Errorf("ethernet1/1/1 = %+v", st)
	}
}

func TestSNMPLLDP(t *testing.T) {
	rem := func(col, local, idx int) string {
		return fmt.Sprintf("%s.%d.0.%d.%d", oidLldpRemTable, col, local, idx)
	}
	w := map[string][]gosnmp.SnmpPDU{
		oidIfName: {strPDU(oidIfName+".1001", "ethernet1/1/1")},
		oidLldpRemTable: {
			intPDU(rem(4, 1001, 1), 4),
			bytesPDU(rem(5, 1001, 1), []byte{0xd0, 0x8e, 0x79, 0, 0, 1}),
			intPDU(rem(6, 1001, 1), 3),
			bytesPDU(rem(7, 1001, 1), []byte{0xd0, 0x8e, 0x79, 0, 0, 1}),
			strPDU(rem(9, 1001, 1), "idrac-ABC1234"),
			strPDU(rem(10, 1001, 1), ""),

			intPDU(rem(4, 54, 1), 4),
			bytesPDU(rem(5, 54, 1), []byte{0x74, 0x83, 0xef, 0, 0, 1}),
			intPDU(rem(6, 54, 1), 5),
			strPDU(rem(7, 54, 1), "Ethernet1/1"),
			strPDU(rem(9, 54, 1), "spine-01"),
			strPDU(rem(10, 54, 1), "Arista Networks EOS"),
		},
		oidLldpLocPortDesc: {strPDU(oidLldpLocPortDesc+".54", "ethernet1/1/54")},
		oidLldpRemManAddrIfIndex: {
			intPDU(oidLldpRemManAddrIfIndex+".0.54.1.1.4.10.0.0.254", 2),
		},
	}

	n := parseSNMPLLDP(w)
	bmc := n["ethernet1/1/1"]
	if bmc == nil || bmc.ChassisIdType != "MAC_ADDRESS" || bmc.ChassisId != "d0:8e:79:00:00:01" || bmc.PortId != "d0:8e:79:00:00:01" || bmc.SystemName != "idrac-ABC1234" {
		t.Errorf("ethernet1/1/1 = %+v", bmc)
	}
	up := n["ethernet1/1/54"]
	if up == nil {
		t.Fatalf("a local port number that is not an ifIndex should be named from lldpLocPortDesc: %v", n)
	}
	if up.PortIdType != "INTERFACE_NAME" || up.PortId != "Ethernet1/1" || up.ManagementAddress != "10.0.0.254" {
		t.Errorf("ethernet1/1/54 = %+v", up)
	}
}

func TestSNMPVersion(t *testing.T) {
	cfg := testConfig()
	cfg.SNMP.Version = "1"
	if _, err := NewSNMP("10.0.0.1", cfg); err == nil {
		t.Error("v1 should be rejected; it has no bulk walk")
	}
}
