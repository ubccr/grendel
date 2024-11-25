package bmc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/ubccr/grendel/util"
)

// PowerCycle will ForceRestart the host
func (r *Redfish) PowerCycle(bootOverride string) error {
	return r.PowerControl(redfish.ForceRestartResetType, bootOverride)
}

// PowerOn will ForceOn the host
func (r *Redfish) PowerOn(bootOverride string) error {
	return r.PowerControl(redfish.OnResetType, bootOverride)
}

// PowerOff will ForceOff the host
func (r *Redfish) PowerOff() error {
	return r.PowerControl(redfish.ForceOffResetType, "")
}

// Power will change the hosts power state
func (r *Redfish) PowerControl(resetType redfish.ResetType, bootOverride string) error {
	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	err = r.bootOverride(bootOverride)
	if err != nil {
		return err
	}

	for _, s := range ss {
		if s.PowerState == redfish.OffPowerState && resetType == redfish.ForceRestartResetType {
			resetType = redfish.OnResetType
		}

		err := s.Reset(resetType)
		if err != nil {
			return err
		}
	}

	return nil
}

// bootOverride will set the boot override target
func (r *Redfish) bootOverride(bootOption string) error {
	if bootOption == "" {
		return nil
	}

	boot := redfish.Boot{
		BootSourceOverrideTarget:  redfish.NoneBootSourceOverrideTarget,
		BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
	}

	switch bootOption {
	case "pxe":
		boot.BootSourceOverrideTarget = redfish.PxeBootSourceOverrideTarget
	case "bios-setup":
		boot.BootSourceOverrideTarget = redfish.BiosSetupBootSourceOverrideTarget
	case "usb":
		boot.BootSourceOverrideTarget = redfish.UsbBootSourceOverrideTarget
	case "hdd":
		boot.BootSourceOverrideTarget = redfish.HddBootSourceOverrideTarget
	case "utilities":
		boot.BootSourceOverrideTarget = redfish.UtilitiesBootSourceOverrideTarget
	case "diagnostics":
		boot.BootSourceOverrideTarget = redfish.DiagsBootSourceOverrideTarget
	case "none":
		return nil
	default:
		return fmt.Errorf("boot option %s not supported", bootOption)
	}

	ss, err := r.service.Systems()
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := s.SetBoot(boot)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Redfish) GetSystem() (*System, error) {
	ss, err := r.service.Systems()
	if err != nil {
		return nil, err
	}

	if len(ss) == 0 {
		return nil, errors.New("failed to find system")
	}

	sys := ss[0]
	var oem SystemOEM
	json.Unmarshal(sys.OEM, &oem)

	system := &System{
		HostName:       sys.HostName,
		BIOSVersion:    sys.BIOSVersion,
		SerialNumber:   sys.SKU,
		Manufacturer:   sys.Manufacturer,
		Model:          sys.Model,
		PowerStatus:    string(sys.PowerState),
		Health:         string(sys.Status.Health),
		TotalMemory:    sys.MemorySummary.TotalSystemMemoryGiB,
		ProcessorCount: sys.ProcessorSummary.LogicalProcessorCount,
		BootNext:       sys.Boot.BootNext,
		BootOrder:      sys.Boot.BootOrder,
		OEM:            oem,
	}

	return system, nil
}
func (r *Redfish) GetFirmware() (map[string]CurrentFirmware, error) {
	us, err := r.service.UpdateService()
	if err != nil {
		return nil, err
	}
	firmwareInventories, err := us.FirmwareInventories()
	if err != nil {
		return nil, err
	}

	layout := "2006-01-02T15:04:05Z"
	firmware := make(map[string]CurrentFirmware, 0)
	for _, firmwareInventory := range firmwareInventories {
		// Dell lists Current, Installed, and Previous firmwares. We only want "Installed" firmware
		if strings.Contains(firmwareInventory.ID, "Previous-") || strings.Contains(firmwareInventory.ID, "Current-") {
			continue
		}
		// Figuring out which version is the newest is a PITA, firmwareInvnetory.VersionSchema may allow us to
		// use the version string, except as of testing, Dell and Lenovo both don't fill out the field
		// So use use the newest released firmware...
		prevReleaseStr := firmware[firmwareInventory.SoftwareID].ReleaseDate
		if prevReleaseStr != "" && prevReleaseStr != "00:00:00Z" {
			prevReleaseTime, err := time.Parse(layout, prevReleaseStr)
			if err != nil {
				return nil, err
			}
			releaseTime, err := time.Parse(layout, firmwareInventory.ReleaseDate)
			if err != nil {
				return nil, err
			}

			if releaseTime.Before(prevReleaseTime) {
				continue
			}
		}
		firmware[firmwareInventory.SoftwareID] = CurrentFirmware{
			ID:          firmwareInventory.ID,
			Name:        firmwareInventory.Name,
			ReleaseDate: firmwareInventory.ReleaseDate,
			SoftwareID:  firmwareInventory.SoftwareID,
			Updatable:   firmwareInventory.Updateable,
			Version:     firmwareInventory.Version,
		}
	}

	return firmware, nil
}

func (r *Redfish) UpdateFirmware(target string) error {
	us, err := r.service.UpdateService()
	if err != nil {
		return err
	}

	rawScheme := viper.GetString("provision.scheme")
	scheme := strings.ToUpper(rawScheme)
	protocol := redfish.HTTPTransferProtocolType
	if scheme == "HTTPS" {
		protocol = redfish.HTTPSTransferProtocolType
	}

	// This should probably be a function?
	rawip, err := util.GetFirstExternalIPFromInterfaces()
	if err != nil {
		return err
	}

	ip := rawip.String()
	lip, port, err := net.SplitHostPort(viper.GetString("provision.listen"))
	if err != nil {
		return err
	}

	if lip != "0.0.0.0" {
		ip = fmt.Sprintf("%s:%s", lip, port)
	}

	cip := viper.GetString("bmc.config_share_ip")
	if cip != "" {
		ip = cip
	}
	//

	params := redfish.SimpleUpdateParameters{
		ForceUpdate: false,
		// Targets: []string{""},
		TransferProtocol: protocol,
		ImageURI:         fmt.Sprintf("%s%s", ip, target),
	}

	return us.SimpleUpdate(&params)
}

func (r *Redfish) GetJobInfo(jid string) (*redfish.Job, error) {
	js, err := r.service.JobService()
	if err != nil {
		return nil, err
	}

	jobs, err := js.Jobs()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.ID == jid {
			return job, nil
		}
	}

	return nil, fmt.Errorf("unable to find job with JID: %s", jid)
}

func (r *Redfish) GetJobs() ([]*redfish.Job, error) {
	js, err := r.service.JobService()
	if err != nil {
		return nil, err
	}

	return js.Jobs()
}
func (r *Redfish) ClearJobs() error {
	js, err := r.service.JobService()
	if err != nil {
		return err
	}

	jobs, err := js.Jobs()
	if err != nil {
		return err
	}

	for _, job := range jobs {
		uri := fmt.Sprintf("/redfish/v1/JobService/Jobs/%s", job.ID)
		resp, err := r.client.Delete(uri)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	return nil
}

func (r *Redfish) GetTaskInfo(tid string) (*redfish.Task, error) {
	ts, err := r.service.TaskService()
	if err != nil {
		return nil, err
	}

	tasks, err := ts.Tasks()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == tid {
			return task, nil
		}
	}

	fmt.Printf("%#v\n", tasks)
	return nil, fmt.Errorf("unable to find task with ID: %s", tid)
}

func (r *Redfish) PowerCycleBmc() error {
	ms, err := r.service.Managers()
	if err != nil {
		return err
	}

	for _, m := range ms {
		err := m.Reset(redfish.GracefulRestartResetType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Redfish) ClearSel() error {
	ms, err := r.service.Managers()
	if err != nil {
		return err
	}

	for _, m := range ms {
		ls, err := m.LogServices()
		if err != nil {
			return err
		}
		for _, l := range ls {
			// ClearLog() errors on other logservice types like "FaultList" on Dell...
			// TODO: Find better solution or add more vendor support below 🙄

			// fmt.Printf("\nID: %s\n Type: %s\n Name: %s\n\n", l.ID, l.LogEntryType, l.Name)
			if l.ID == "Sel" || l.ID == "Log1" || len(ls) == 1 {
				err := l.ClearLog()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *Redfish) BmcAutoConfigure() error {
	// TODO: Submit PR to gofish to support this natively?

	type attributes struct {
		NIC1AutoConfig string `json:"NIC.1.AutoConfig"`
	}
	type payload struct {
		Attributes attributes
	}

	p := payload{
		Attributes: attributes{
			NIC1AutoConfig: "Enable Once",
		},
	}

	return r.service.Patch("/redfish/v1/Managers/iDRAC.Embedded.1/Attributes", p)
}

func (r *Redfish) BmcImportConfiguration(shutdownType, path, file string) (string, error) {
	// TODO: Submit PR to gofish to support this natively?

	type shareParameters struct {
		Target                   []string
		ShareType                string
		IPAddress                string
		FileName                 string
		ShareName                string
		PortNumber               string
		IgnoreCertificateWarning string
	}
	type payload struct {
		HostPowerState  string
		ShutdownType    string
		ImportBuffer    string
		ShareParameters shareParameters
	}

	// viper.SetDefault("provision.scheme", "http")
	rawScheme := viper.GetString("provision.scheme")
	scheme := strings.ToUpper(rawScheme)

	rawip, err := util.GetFirstExternalIPFromInterfaces()
	if err != nil {
		return "", err
	}

	ip := rawip.String()
	lip, port, err := net.SplitHostPort(viper.GetString("provision.listen"))
	if err != nil {
		return "", err
	}

	if lip != "0.0.0.0" {
		ip = lip
	}

	cip := viper.GetString("bmc.config_share_ip")
	if cip != "" {
		ip = cip
	}

	// viper.SetDefault("bmc.config_ignore_certificate_warning", "Disabled")
	icw := viper.GetString("bmc.config_ignore_certificate_warning")

	log.Debugf("Import system config debug info: scheme=%s ip=%s port=%s file=%s path=%s icw=%s", scheme, ip, port, file, path, icw)

	p := payload{
		HostPowerState: "On",
		ShutdownType:   shutdownType,
		ShareParameters: shareParameters{
			Target:                   []string{"ALL"},
			ShareType:                scheme,
			IPAddress:                ip,
			PortNumber:               port,
			FileName:                 file,
			ShareName:                path,
			IgnoreCertificateWarning: icw,
		},
	}

	err = r.service.Post("/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration", p)
	if err != nil {
		return "", err
	}

	// Get job ID
	j, err := r.service.JobService()
	if err != nil {
		return "", err
	}

	jobs, err := j.Jobs()
	if err != nil {
		return "", err
	}

	for _, job := range jobs {
		if job.Name == "Import Configuration" && job.JobState == redfish.RunningJobState {
			return job.ID, nil
		}
	}
	return "", errors.New("failed to find job")
}

func (r *Redfish) BmcGetJob(id string) (*redfish.Job, error) {
	j, err := r.service.JobService()
	if err != nil {
		return nil, err
	}

	jobs, err := j.Jobs()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.ID == id {
			return job, nil
		}
	}
	return nil, errors.New("failed to find job")
}
