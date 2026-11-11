// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	idracManagerPath = "/redfish/v1/Managers/iDRAC.Embedded.1"
	importTarget     = idracManagerPath + "/Actions/Oem/EID_674_Manager.ImportSystemConfiguration"
	sisPath          = "/redfish/v1/Dell/Systems/System.Embedded.1/DellSoftwareInstallationService"
	installTarget    = sisPath + "/Actions/DellSoftwareInstallationService.InstallFromRepository"
	updateListTarget = sisPath + "/Actions/DellSoftwareInstallationService.GetRepoBasedUpdateList"
	dellJobPath      = "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/Jobs/JID_9"
)

type sentRequest struct {
	method, path string
	body         map[string]any
}

// idracFixture serves an iDRAC with the OEM actions these operations use. Every request that is not a GET is recorded in order, and jobStates is what successive reads of the job report
func idracFixture(t *testing.T, jobStates ...string) (*fixture, *[]sentRequest) {
	t.Helper()

	f := dellFixture()
	f.tree["/redfish/v1/Managers"] = `{
	  "@odata.id": "/redfish/v1/Managers",
	  "Members@odata.count": 1,
	  "Members": [{"@odata.id": "` + idracManagerPath + `"}]
	}`
	f.tree[idracManagerPath] = `{
	  "@odata.id": "` + idracManagerPath + `",
	  "@odata.type": "#Manager.v1_5_0.Manager",
	  "Id": "iDRAC.Embedded.1",
	  "Name": "Manager",
	  "Actions": {"Oem": {"#OemManager.ImportSystemConfiguration": {"target": "` + importTarget + `"}}}
	}`
	f.tree["/redfish/v1/Systems/System.Embedded.1"] = strings.Replace(systemDell,
		`"Oem": {"Dell"`,
		`"Links": {"Oem": {"Dell": {"DellSoftwareInstallationService": {"@odata.id": "`+sisPath+`"}}}},
	  "Oem": {"Dell"`, 1)
	f.tree[sisPath] = `{
	  "@odata.id": "` + sisPath + `",
	  "Id": "DellSoftwareInstallationService",
	  "Actions": {
	    "#DellSoftwareInstallationService.InstallFromRepository": {"target": "` + installTarget + `"},
	    "#DellSoftwareInstallationService.GetRepoBasedUpdateList": {"target": "` + updateListTarget + `"}
	  }
	}`
	f.tree["/redfish/v1/TaskService/Tasks/JID_9"] = `{
	  "@odata.id": "/redfish/v1/TaskService/Tasks/JID_9",
	  "@odata.type": "#Task.v1_4_0.Task",
	  "Id": "JID_9",
	  "Name": "Import Configuration"
	}`

	sent := &[]sentRequest{}
	reads := 0
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodGet && r.URL.Path == dellJobPath {
			state := "Completed"
			if len(jobStates) > 0 {
				state = jobStates[min(reads, len(jobStates)-1)]
			}
			reads++
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"@odata.id": %q, "Id": "JID_9", "JobState": %q, "Message": "catalog unreachable", "PercentComplete": 0}`, dellJobPath, state)
			return true
		}
		if r.Method == http.MethodGet {
			return false
		}

		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		*sent = append(*sent, sentRequest{method: r.Method, path: r.URL.Path, body: body})

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case importTarget:
			w.Header().Set("Location", "/redfish/v1/TaskService/Tasks/JID_9")
			w.WriteHeader(http.StatusAccepted)
		case installTarget:
			w.Header().Set("Location", dellJobPath)
			w.WriteHeader(http.StatusAccepted)
		case updateListTarget:
			fmt.Fprintf(w, `{"PackageList": %q}`, packageListXML)
		default:
			w.WriteHeader(http.StatusOK)
		}
		return true
	}

	return f, sent
}

const packageListXML = `<CIM><MESSAGE><SIMPLEREQ>
<VALUE.NAMEDINSTANCE><INSTANCENAME>
<PROPERTY NAME="DisplayName"><VALUE>BIOS</VALUE></PROPERTY>
<PROPERTY NAME="PackageVersion"><VALUE>2.20.0</VALUE></PROPERTY>
<PROPERTY NAME="RebootType"><VALUE>HOST</VALUE></PROPERTY>
<PROPERTY.ARRAY NAME="ComponentInstalledVersion"><VALUE.ARRAY><VALUE>2.19.1</VALUE></VALUE.ARRAY></PROPERTY.ARRAY>
</INSTANCENAME></VALUE.NAMEDINSTANCE>
</SIMPLEREQ></MESSAGE></CIM>`

func idracTarget(f *fixture, t *testing.T) []Target {
	return []Target{{Name: "cpn-a01-01", Addr: f.serve(t).URL}}
}

func TestAutoConfigurePatchesAttributes(t *testing.T) {
	f, sent := idracFixture(t)

	r := testService(nil).AutoConfigure(context.Background(), idracTarget(f, t))[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, "set NIC.1.AutoConfig to Enable Once", r.Message)
	require.Len(t, *sent, 1)
	assert.Equal(t, http.MethodPatch, (*sent)[0].method)
	assert.Equal(t, idracManagerPath+"/Attributes", (*sent)[0].path)
	assert.Equal(t, map[string]any{"NIC.1.AutoConfig": "Enable Once"}, (*sent)[0].body["Attributes"])
}

// TestDellOperationsRefuseOtherVendors checks that nothing is sent to a BMC that is not an iDRAC, since every one of these is a Dell OEM action
func TestDellOperationsRefuseOtherVendors(t *testing.T) {
	f, sent := idracFixture(t)
	f.tree["/redfish/v1/"] = serviceRootGeneric
	targets := idracTarget(f, t)
	svc := testService(nil)
	ctx := context.Background()

	results := []NodeResult[Ack]{
		svc.AutoConfigure(ctx, targets)[0],
		svc.ImportConfig(ctx, targets, ImportConfigRequest{ShareNames: map[string]string{"cpn-a01-01": "/boot/x/bmc"}})[0],
		svc.InstallFirmware(ctx, targets, InstallFirmwareRequest{})[0],
	}
	for _, r := range results {
		assert.False(t, r.Success)
		assert.Contains(t, r.Message, `only supported on a Dell iDRAC, this bmc reports vendor "Contoso"`)
	}

	updates := svc.FirmwareUpdates(ctx, targets)[0]
	assert.False(t, updates.Success)
	assert.Empty(t, *sent)
}

func TestImportConfigSendsTheNodesShare(t *testing.T) {
	f, sent := idracFixture(t)

	req := ImportConfigRequest{
		Body: dell.ImportSystemConfigurationBody{
			ShutdownType: dell.NoRebootISCShutdownType,
			ShareParameters: dell.ShareParameters{
				Target:    "ALL",
				ShareType: dell.HTTPISCShareType,
				IPAddress: "10.0.0.1",
				FileName:  "idrac.json",
			},
		},
		ShareNames: map[string]string{"cpn-a01-01": "/boot/token-a01/bmc"},
	}

	r := testService(nil).ImportConfig(context.Background(), idracTarget(f, t), req)[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, "submitted import job JID_9", r.Message)
	require.Len(t, *sent, 1)
	assert.Equal(t, importTarget, (*sent)[0].path)
	share := (*sent)[0].body["ShareParameters"].(map[string]any)
	assert.Equal(t, "/boot/token-a01/bmc", share["ShareName"])
	assert.Equal(t, "idrac.json", share["FileName"])
	assert.Equal(t, "10.0.0.1", share["IPAddress"])
}

// TestImportConfigSendsThePort checks the port goes out only when it is asked for, since iDRAC firmware before 7.00.00 does not take it
func TestImportConfigSendsThePort(t *testing.T) {
	for _, port := range []string{"8080", ""} {
		f, sent := idracFixture(t)
		req := ImportConfigRequest{
			Body:       dell.ImportSystemConfigurationBody{ShareParameters: dell.ShareParameters{Target: "ALL", ShareType: dell.HTTPISCShareType}},
			PortNumber: port,
			ShareNames: map[string]string{"cpn-a01-01": "/boot/token-a01/bmc"},
		}

		r := testService(nil).ImportConfig(context.Background(), idracTarget(f, t), req)[0]

		require.True(t, r.Success, "message was: %s", r.Message)
		share := (*sent)[0].body["ShareParameters"].(map[string]any)
		got, present := share["PortNumber"]
		if port == "" {
			assert.False(t, present, "an empty port must be left out, got %v", got)
		} else {
			assert.Equal(t, port, got)
		}
		assert.Equal(t, "/boot/token-a01/bmc", share["ShareName"])
	}
}

func TestImportConfigWithoutAShareFailsTheNode(t *testing.T) {
	f, sent := idracFixture(t)

	r := testService(nil).ImportConfig(context.Background(), idracTarget(f, t), ImportConfigRequest{})[0]

	assert.False(t, r.Success)
	assert.Contains(t, r.Message, "no share path")
	assert.Empty(t, *sent)
}

func TestFirmwareUpdatesParsesThePackageList(t *testing.T) {
	f, _ := idracFixture(t)

	r := testService(nil).FirmwareUpdates(context.Background(), idracTarget(f, t))[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	require.Len(t, r.Data.Updates, 1)
	u := r.Data.Updates[0]
	assert.Equal(t, "BIOS", u.DisplayName)
	assert.Equal(t, "2.19.1", u.InstalledVersion)
	assert.Equal(t, "2.20.0", u.PackageVersion)
	assert.Equal(t, "HOST", u.RebootType)
}

func installBody(apply dell.ApplyUpdate) dell.InstallFromRepoBody {
	return dell.InstallFromRepoBody{ApplyUpdate: apply, IPAddress: "downloads.dell.com", ShareName: "/", ShareType: dell.HTTPSIFRShareType}
}

// TestInstallFirmwareClearsTheQueueFirst checks the order on the wire: the old jobs are gone before the update is queued
func TestInstallFirmwareClearsTheQueueFirst(t *testing.T) {
	f, sent := idracFixture(t)
	jobs, _ := jobFixture([]string{"JID_1"})
	for path, body := range jobs.tree {
		if strings.HasPrefix(path, jobServicePath) {
			f.tree[path] = body
		}
	}

	r := testService(nil).InstallFirmware(context.Background(), idracTarget(f, t),
		InstallFirmwareRequest{Body: installBody(dell.ApplyUpdateTrue), ClearJobs: true})[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, "queued firmware update job JID_9", r.Message)
	require.Len(t, *sent, 2)
	assert.Equal(t, sentRequest{method: http.MethodDelete, path: jobServicePath + "/Jobs/JID_1"}, sentRequest{method: (*sent)[0].method, path: (*sent)[0].path})
	assert.Equal(t, installTarget, (*sent)[1].path)
}

// TestInstallFirmwareCheckWaitsForTheJob is a comparison without ApplyUpdate, which is only useful once the job is done
func TestInstallFirmwareCheckWaitsForTheJob(t *testing.T) {
	firmwarePollInterval = time.Millisecond
	t.Cleanup(func() { firmwarePollInterval = 5 * time.Second })

	f, sent := idracFixture(t, "Running", "Running", "Completed")

	r := testService(nil).InstallFirmware(context.Background(), idracTarget(f, t),
		InstallFirmwareRequest{Body: installBody(dell.ApplyUpdateFalse), ClearJobs: true})[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Contains(t, r.Message, "compared firmware against the catalog in job JID_9")
	require.Len(t, *sent, 1, "clear jobs only applies when the update is applied")
	assert.Equal(t, installTarget, (*sent)[0].path)
}

func TestInstallFirmwareCheckReportsAFailedJob(t *testing.T) {
	firmwarePollInterval = time.Millisecond
	t.Cleanup(func() { firmwarePollInterval = 5 * time.Second })

	f, _ := idracFixture(t, "Running", "Failed")

	r := testService(nil).InstallFirmware(context.Background(), idracTarget(f, t),
		InstallFirmwareRequest{Body: installBody(dell.ApplyUpdateFalse)})[0]

	assert.False(t, r.Success)
	assert.Contains(t, r.Message, "job JID_9 ended Failed: catalog unreachable")
}

func TestInstallFirmwareCheckStopsAtTheNodeTimeout(t *testing.T) {
	firmwarePollInterval = time.Millisecond
	t.Cleanup(func() { firmwarePollInterval = 5 * time.Second })

	f, _ := idracFixture(t, "Running")
	svc := testService(func(c *Config) { c.NodeTimeout = 200 * time.Millisecond })

	r := svc.InstallFirmware(context.Background(), idracTarget(f, t),
		InstallFirmwareRequest{Body: installBody(dell.ApplyUpdateFalse)})[0]

	assert.False(t, r.Success)
}
