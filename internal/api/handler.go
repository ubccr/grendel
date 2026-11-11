// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/eventstore"
)

type Handler struct {
	DB     store.Store
	Events eventstore.Store
	// Redfish serves the bmc routes
	Redfish *redfish.Service
}

func NewHandler(db store.Store) (*Handler, error) {
	h := &Handler{DB: db, Redfish: redfish.NewFromViper()}

	return h, nil
}

func (h *Handler) SetupRoutes(s *fuego.Server) {
	if viper.GetBool("api.web_ui") {
		fuego.Handle(s, "/ui/", setupFrontend())
		fuego.Handle(s, "/{$}", http.RedirectHandler("/ui", http.StatusMovedPermanently))
	}

	// Examples
	nsExample := param.Example("nodeset", "cpn-i10-[04-05],cpn-h22-33")
	usernamesExample := option.Path("usernames", "target usernames", param.Example("usernames", "user1,user2"))

	// Params
	filterNodes := fuego.GroupOptions(option.Query("nodeset", "Filter by nodeset. Minimum of one query parameter is required", nsExample), option.Query("tags", "Filter by tags. Minimum of one query parameter is required", param.Example("tags", "a01,ib,test")))
	selectAll := option.Query("all", "Select every node in the database. Required when no nodeset or tags are given", param.Example("all", "true"))
	filterNames := fuego.GroupOptions(option.Query("names", "Filter by name", param.Example("names", "image1,image2")))

	globalOptions := fuego.GroupOptions(
		option.RequestContentType("application/json"),
		fuego.OptionRemoveResponse(400),
		fuego.OptionRemoveResponse(500),
		fuego.OptionAddDefaultResponse("Default Error", fuego.Response{Type: fuego.HTTPError{}, ContentTypes: []string{"application/json"}}),
	)

	// Groups
	v1 := fuego.Group(s, "/v1",
		option.Security(
			openapi3.NewSecurityRequirement().Authenticate("headerAuth"),
			openapi3.NewSecurityRequirement().Authenticate("cookieAuth"),
		),
	)
	grendel := fuego.Group(v1, "/grendel", option.Middleware(h.authMiddleware), globalOptions)
	nodes := fuego.Group(v1, "/nodes", option.Middleware(h.authMiddleware), globalOptions)
	images := fuego.Group(v1, "/images", option.Middleware(h.authMiddleware), globalOptions)
	users := fuego.Group(v1, "/users", option.Middleware(h.authMiddleware), globalOptions)
	auth := fuego.Group(v1, "/auth", globalOptions)
	db := fuego.Group(v1, "/db", option.Middleware(h.authMiddleware), globalOptions)
	bmc := fuego.Group(v1, "/bmc", option.Middleware(h.authMiddleware), globalOptions)
	roles := fuego.Group(v1, "/roles", option.Middleware(h.authMiddleware), globalOptions)
	sw := fuego.Group(v1, "/switch", option.Middleware(h.authMiddleware), globalOptions)

	// Routes
	fuego.Get(grendel, "/events", h.GetEvents)
	fuego.Get(grendel, "/version", h.GetVersion, option.Description("Get the server version"))

	fuego.Post(nodes, "", h.NodeAdd, option.Description("Add nodes"))
	fuego.Get(nodes, "", h.NodeList, option.Description("List all nodes"))
	fuego.Delete(nodes, "", h.NodeDelete,
		option.Description("Delete nodes by nodeset and/or tags"),
		filterNodes,
	)
	fuego.Get(nodes, "/find", h.NodeFind,
		option.Description("Find nodes by nodeset and/or tags"),
		filterNodes,
	)
	fuego.Patch(nodes, "/provision", h.NodeProvision,
		option.Description("Provision / Unprovision nodes by nodeset and/or tags"),
		filterNodes,
	)
	fuego.Patch(nodes, "/tags/{action}", h.NodeTags,
		option.Description("Update nodes tags by nodeset and/or tags"),
		option.Path("action", "option to add or remove tags", param.Example("action", "add | remove")),
		filterNodes,
	)
	fuego.Get(nodes, "/token/{interface}", h.NodeBootToken,
		option.Description("Create a boot token for the provision server. Used for debugging requests made by images"),
		option.Path("interface", "interface token will be created for", param.Example("interface", "boot | bmc")),
		filterNodes,
		option.Query("arch", "Arch to embed in the token, used in templates when the node has no arch set. Simulates the arch detected during a PXE boot", param.Example("arch", "x86_64")),
	)
	fuego.Patch(nodes, "/image", h.NodeBootImage,
		option.Description("Update nodes boot image by nodeset and/or tags"),
		filterNodes,
	)

	fuego.Post(images, "", h.BootImageAdd, option.Description("Add images"))
	fuego.Get(images, "", h.BootImageList, option.Description("List all images"))
	fuego.Delete(images, "", h.BootImageDelete, option.Description("Delete images by name"), filterNames)
	fuego.Get(images, "/find", h.BootImageFind, option.Description("Find images by name"), filterNames)

	fuego.Post(users, "", h.UserStore, option.Description("Add new user"))
	fuego.Get(users, "", h.UserList, option.Description("List all users"), option.Query("usernames", "Filter by usernames", param.Example("username", "admin,user")))
	fuego.Delete(users, "/{usernames}", h.UserDelete,
		option.Description("Delete users"),
		usernamesExample,
	)
	fuego.Patch(users, "/{usernames}/role", h.UserRole,
		option.Description("Update users role"),
		usernamesExample,
	)
	fuego.Patch(users, "/{usernames}/enable", h.UserEnable,
		option.Description("Update users enable"),
		usernamesExample,
	)

	fuego.Post(auth, "/signin", h.AuthSignin,
		option.Description("signin user"),
		option.Security(openapi3.NewSecurityRequirement()),
	)
	fuego.Post(auth, "/signup", h.AuthSignup,
		option.Description("Signup user"),
		option.Security(openapi3.NewSecurityRequirement()),
	)
	fuego.Delete(auth, "/signout", h.AuthSignout,
		option.Description("Signout user"),
		option.Security(openapi3.NewSecurityRequirement()),
	)
	fuego.Post(auth, "/token", h.AuthToken,
		option.Description("Create API token"),
		option.Middleware(h.authMiddleware),
	)
	fuego.Patch(auth, "/reset", h.AuthReset,
		option.Description("Change password"),
		option.Middleware(h.authMiddleware),
	)

	fuego.Post(db, "/restore", h.Restore, option.Description("Restore a backup of the DB"))
	fuego.Get(db, "/dump", h.Dump, option.Description("Get a backup of the DB"))

	fuego.Get(bmc, "/status", h.BmcStatus,
		option.Description("Get redfish system information from node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/power", h.BmcPower,
		option.Description("Change the power state of node(s), or restart their BMCs"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/indicator", h.BmcIndicator,
		option.Description("Turn the location indicator of node(s) on or off"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/bios", h.BmcBios,
		option.Description("Get the BIOS attributes of node(s), with any staged changes"),
		option.Query("names", "Keep only these attributes", param.Example("names", "SriovGlobalEnable,ProcCStates")),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/bios", h.BmcSetBios,
		option.Description("Stage BIOS attribute changes on node(s), applied at the given apply time"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/bios/discard", h.BmcDiscardBios,
		option.Description("Discard the BIOS changes staged on node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/bios/registry", h.BmcBiosRegistry,
		option.Description("Get the attribute registry describing one node's BIOS attributes"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/jobs", h.BmcJobs,
		option.Description("Get the redfish job queue of node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/jobs/clear", h.BmcClearJobs,
		option.Description("Clear redfish jobs from node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/sel", h.BmcSEL,
		option.Description("Get the system event log of node(s)"),
		option.Query("limit", "Keep only this many entries per node", param.Example("limit", "50")),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/sel/clear", h.BmcClearSEL,
		option.Description("Clear the system event log on node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/configure/auto", h.BmcAutoConfigure,
		option.Description("Set iDRAC(s) to auto configure once, on the node's next boot. Dell only"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/configure/import", h.BmcImportConfig,
		option.Description("Have iDRAC(s) download a config template from the provision server and import it. Dell only"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/firmware/updates", h.BmcFirmwareUpdates,
		option.Description("List the firmware updates iDRAC(s) found in their last catalog comparison, which firmware/install with apply_update false runs. Dell only"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/firmware/install", h.BmcInstallFirmware,
		option.Description("Compare the firmware of iDRAC(s) against an update catalog, and with apply_update queue the updates. A comparison waits for its job to finish, up to bmc.node_timeout. Dell only"),
		filterNodes,
		selectAll,
	)

	fuego.Get(sw, "/{nodeset}/lldp", h.SwitchGetLLDP,
		option.Description("Get switch LLDP info"),
		option.Query("ports", "Filter by port name", param.Example("ports", "Et1,Et2")),
		// filterNodes,
	)

	fuego.Get(roles, "", h.GetRoles,
		option.Description("Get roles and permissions"),
		option.Query("name", "Filter by name", param.Example("name", "admin,user")),
	)
	fuego.Patch(roles, "", h.PatchRoles,
		option.Description("Edit role permissions"),
	)
	fuego.Post(roles, "", h.PostRoles,
		option.Description("Add roles"),
	)
	fuego.Delete(roles, "/{names}", h.DeleteRoles,
		option.Description("Delete roles"),
		option.Path("names", "Delete by name", param.Example("names", "admin,user")),
	)

}
