// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/client"
	"github.com/ubccr/grendel/pkg/model"
)

func TestShutdownType(t *testing.T) {
	cases := map[string]dell.ISCShutdownType{
		"":         dell.NoRebootISCShutdownType,
		"NoReboot": dell.NoRebootISCShutdownType,
		"Graceful": dell.GracefulISCShutdownType,
		"Forced":   dell.ForcedISCShutdownType,
	}
	for raw, want := range cases {
		got, err := shutdownType(raw)
		if err != nil || got != want {
			t.Errorf("shutdownType(%q) = %q, %v, want %q", raw, got, err, want)
		}
	}

	// v1 quietly turned an unknown value into NoReboot, which hid a typo of Forced
	if _, err := shutdownType("forced"); err == nil {
		t.Error("an unknown shutdown type must be rejected")
	}
}

func TestInstallFirmwareRequest(t *testing.T) {
	valid := BmcInstallFirmwareBody{IPAddress: "downloads.dell.com", ShareType: "https", ShareName: "/"}

	req, err := installFirmwareRequest(valid)
	if err != nil {
		t.Fatal(err)
	}
	if req.Body.ShareType != dell.HTTPSIFRShareType {
		t.Errorf("share type = %q, the case should not matter", req.Body.ShareType)
	}
	if req.Body.ApplyUpdate != dell.ApplyUpdateFalse || req.Body.IgnoreCertWarning != dell.IgnoreCertWarningOff {
		t.Errorf("the booleans did not map, got %+v", req.Body)
	}

	cases := map[string]func(b *BmcInstallFirmwareBody){
		"no ip address":             func(b *BmcInstallFirmwareBody) { b.IPAddress = "" },
		"no share name":             func(b *BmcInstallFirmwareBody) { b.ShareName = "" },
		"cifs, which needs a login": func(b *BmcInstallFirmwareBody) { b.ShareType = "CIFS" },
		"unknown share type":        func(b *BmcInstallFirmwareBody) { b.ShareType = "FTP" },
		"clear jobs on a check":     func(b *BmcInstallFirmwareBody) { b.ClearJobs = true },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			body := valid
			mutate(&body)
			if _, err := installFirmwareRequest(body); err == nil {
				t.Error("expected the request to be rejected")
			}
		})
	}
}

func TestImportShare(t *testing.T) {
	v := viper.New()
	v.Set("provision.listen", "10.1.0.5:80")
	share, err := importShare(v)
	if err != nil {
		t.Fatal(err)
	}
	if share.ip != "10.1.0.5" || share.port != "" || share.shareType != dell.HTTPISCShareType || share.ignoreCert != dell.DisabledISCIgnoreCertificateWarning {
		t.Errorf("listen address share = %+v", share)
	}

	v.Set("bmc.config_share_ip", "192.168.0.9")
	v.Set("provision.cert", "/etc/grendel/cert.pem")
	v.Set("bmc.config_ignore_certificate_warning", "Enabled")
	share, err = importShare(v)
	if err != nil {
		t.Fatal(err)
	}
	// port 80 is not the https default, so it has to be sent
	if share.ip != "192.168.0.9" || share.port != "80" || share.shareType != dell.HTTPSISCShareType || share.ignoreCert != dell.EnabledISCIgnoreCertificateWarning {
		t.Errorf("override share = %+v", share)
	}

	v.Set("provision.listen", "0.0.0.0:443")
	share, err = importShare(v)
	if err != nil {
		t.Fatal(err)
	}
	if share.port != "" {
		t.Errorf("the https default port must not be sent, got %q", share.port)
	}
}

func TestShareNames(t *testing.T) {
	withBMC := tests.HostFactory.MustCreate().(*model.Host)
	noBMC := tests.HostFactory.MustCreate().(*model.Host)
	for i := range noBMC.Interfaces {
		noBMC.Interfaces[i].BMC = false
	}

	names := shareNames(model.HostList{withBMC, noBMC, nil})

	if len(names) != 1 {
		t.Fatalf("names = %v, want only the node with a bmc interface", names)
	}
	path := names[withBMC.Name]
	if !strings.HasPrefix(path, "/boot/") || !strings.HasSuffix(path, "/bmc") || strings.Count(path, "/") != 3 {
		t.Errorf("share path = %q, want /boot/<token>/bmc", path)
	}
}

func TestFirmwareUpdatesRebootType(t *testing.T) {
	cases := []struct {
		reboots []string
		want    string
	}{
		{nil, "NONE"},
		{[]string{"NONE", "IDRAC"}, "IDRAC"},
		{[]string{"HOST", "IDRAC", "NONE"}, "HOST"},
		{[]string{"IDRAC", "HOST"}, "HOST"},
	}
	for _, c := range cases {
		var data redfish.FirmwareUpdates
		for _, r := range c.reboots {
			data.Updates = append(data.Updates, dell.UpdateListFirmware{RebootType: r})
		}
		if got := firmwareUpdates(&data).RebootType; got != c.want {
			t.Errorf("reboot type of %v = %q, want %q", c.reboots, got, c.want)
		}
	}

	if firmwareUpdates(nil) != nil {
		t.Error("a failed node has no data")
	}
}

func TestGeneratedClientDecodesFirmwareUpdates(t *testing.T) {
	body, err := json.Marshal(model.BmcResponse[model.BmcFirmwareUpdates]{RunID: "run", Results: []model.BmcResult[model.BmcFirmwareUpdates]{
		{Node: "cpn-1", Success: true, Data: firmwareUpdates(&redfish.FirmwareUpdates{Updates: dell.UpdateList{{DisplayName: "BIOS", InstalledVersion: "2.19.1", PackageVersion: "2.20.0", RebootType: "HOST"}}})},
		{Node: "cpn-2", Message: "only supported on a Dell iDRAC"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	var res client.BmcResponseModelBmcFirmwareUpdates
	if err := res.Decode(jx.DecodeBytes(body)); err != nil {
		t.Fatalf("decoding %s: %v", body, err)
	}

	data := res.Results.Value[0].Value.Data.Value
	if data.RebootType.Value != "HOST" || data.Updates.Value[0].Value.PackageVersion.Value != "2.20.0" {
		t.Errorf("decoded %+v", data)
	}
	if !res.Results.Value[1].Value.Data.Null {
		t.Error("a failed node should decode null data")
	}
}
