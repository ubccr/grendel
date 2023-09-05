package frontend

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/tors"
)

func GetMacAddress(h *Handler, f *fiber.Ctx, rack string, switchType string, hosts []string) (map[string]string, error) {
	switchTag := "mgmt_leaf"
	if switchType == "Core" {
		switchTag = "core_leaf"
	}
	hostMacMap := make(map[string]string)

	nodeset, err := h.DB.MatchTags([]string{rack, switchTag})
	if err != nil {
		return hostMacMap, err
	}
	host, err := h.DB.FindHosts(nodeset)
	if err != nil {
		return hostMacMap, err
	}

	endpoint := fmt.Sprintf("https://%s", host[0].InterfaceBMC().ToStdAddr().String())
	sw, err := tors.NewDellOS10(endpoint, "admin", viper.GetString("bmc.switch_admin_password"), "", true)
	if err != nil {
		return hostMacMap, err
	}

	macTable, err := sw.GetMACTable()
	if err != nil {
		return hostMacMap, err
	}

	for _, v := range hosts {
		portString := f.FormValue(fmt.Sprintf("%s:%s", v, switchType))
		port, err := strconv.Atoi(portString)
		if err != nil {
			return hostMacMap, err
		}

		rawMac := macTable.Port(port)
		mac := rawMac[0].MAC.String()

		hostMacMap[v] = mac
	}
	return hostMacMap, nil
}
