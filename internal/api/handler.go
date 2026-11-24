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
	"github.com/ubccr/grendel/internal/events"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
)

type Handler struct {
	DB      store.Store
	Events  *events.Recorder
	Redfish *redfish.Service
	// routes describes each registered route by its permission, for the role list
	routes map[model.Permission]routeInfo
}

func NewHandler(db store.Store) (*Handler, error) {
	h := &Handler{DB: db, Events: events.New(db), Redfish: redfish.NewFromSettings()}

	return h, nil
}

func (h *Handler) SetupRoutes(s *fuego.Server) {
	if settings.APIWebUI.Get() {
		fuego.Handle(s, "/ui/", setupFrontend())
		fuego.Handle(s, "/{$}", http.RedirectHandler("/ui", http.StatusMovedPermanently))
	}

	// Examples
	nsExample := param.Example("nodeset", "cpn-i10-[04-05],cpn-h22-33")
	usernamesExample := option.Path("usernames", "target usernames", param.Example("usernames", "user1,user2"))

	// Params
	filterNodes := fuego.GroupOptions(option.Query("nodeset", "Filter by nodeset. Minimum of one query parameter is required", nsExample), option.Query("tags", "Filter by tags. Minimum of one query parameter is required", param.Example("tags", "a01,ib,test")))
	selectAll := option.QueryBool("all", "Select every node in the database. Required when no nodeset or tags are given", param.Default(false))
	filterNames := fuego.GroupOptions(option.Query("names", "Filter by name", param.Example("names", "image1,image2")))

	renderOptions := fuego.GroupOptions(
		option.Query("arch", "Arch for nodes with no arch set, standing in for the one detected during a PXE boot. With all, such a node is taken once as x86_64 and once as aarch64", param.Example("arch", "x86_64 | aarch64 | all")),
		option.Query("server", "Address of the provision server to put in URLs. Defaults to the address provision listens on, or the machine's first external address, and provision.hostname replaces either", param.Example("server", "10.17.0.1")),
	)

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
	config := fuego.Group(v1, "/settings", option.Middleware(h.authMiddleware), globalOptions)

	// Routes
	fuego.Get(grendel, "/events", h.GetEvents, about("Follow events", "Server"),
		option.Description("List the latest events, newest first, with how many of their nodes are at each state"),
		option.Query("state", "Only events at this state", param.Example("state", "running | done")),
		option.Query("kind", "Only events of this kind", param.Example("kind", "audit | bmc | provision | firmware")),
		option.Query("before", "Only events older than the event with this id, for reading further back"),
		option.QueryInt("limit", "The most events to return, never more than 1000", param.Default(model.DefaultEventLimit)),
	)
	fuego.Get(grendel, "/events/{id}", h.GetEvent, about("See how an event went on each node", "Server"),
		option.Description("Get one event with how it went on each of its nodes"),
		option.Path("id", "id of the event"),
	)
	fuego.Get(grendel, "/version", h.GetVersion, about("See the server version", "Server"), option.Description("Get the server version"))

	fuego.Post(nodes, "", h.NodeAdd, about("Add and edit nodes", "Nodes"), option.Description("Add nodes"))
	fuego.Get(nodes, "", h.NodeList, about("List nodes", "Nodes"), option.Description("List all nodes"))
	fuego.Delete(nodes, "", h.NodeDelete, about("Delete nodes", "Nodes"),
		option.Description("Delete nodes by nodeset and/or tags"),
		filterNodes,
	)
	fuego.Get(nodes, "/find", h.NodeFind, about("Find nodes", "Nodes"),
		option.Description("Find nodes by nodeset and/or tags"),
		filterNodes,
	)
	fuego.Patch(nodes, "/provision", h.NodeProvision, about("Turn provisioning on or off", "Provisioning"),
		option.Description("Provision / Unprovision nodes by nodeset and/or tags"),
		filterNodes,
	)
	fuego.Get(nodes, "/provision-status", h.NodeProvisionStatus, about("See provision status", "Provisioning"),
		option.Description("Get the latest provision status of nodes by nodeset and/or tags, or of every node when neither is given"),
		filterNodes,
	)
	fuego.Get(nodes, "/tags", h.NodeTagList, about("List tags", "Nodes"),
		option.Description("List every tag on a node, with how many nodes have it"),
	)
	fuego.Post(nodes, "/import", h.NodeImport, about("Import nodes from switches and BMCs", "Nodes"),
		option.Description("Add a BMC or boot port to each node, creating nodes that do not exist, with addresses taken from ranges. A dry run reports what would change; a commit stores every row or none"),
	)
	fuego.Get(nodes, "/ips/next", h.NodeNextAddress, about("Find free addresses", "Addresses"),
		option.Description("Suggest free addresses in a range: the lowest ones no node holds, and the ones after the highest held. Suggestions are pinged, and one that answers is listed but marked"),
		option.Query("cidr", "The IPv4 range to search, a /16 or smaller: a CIDR, an address standing for its /24, or three octets like 10.17.25", param.Required(), param.Example("cidr", "10.17.25.0/24")),
		option.Query("from", "Suggest only the lowest free addresses, or only those after the highest used. Both lists are returned without it", param.Example("from", "lowest | after_highest")),
		option.QueryInt("count", "How many unanswered addresses to suggest for each list, from 1 to 8 when pinging or to 64 without", param.Default(defaultNextCount)),
		option.QueryBool("ping", "Ping the suggestions", param.Default(true)),
		option.Query("exclude", "Addresses to treat as taken, like ones on unsaved ports, with or without a prefix length", param.Example("exclude", "10.17.25.7/18,10.17.25.8")),
	)
	fuego.Get(nodes, "/ips/ranges", h.NodeAddressRanges, about("List address ranges in use", "Addresses"),
		option.Description("List the ranges the ports of some nodes sit in, each once however many ports share it: a /24 of a larger subnet, or a /24 to /30 subnet whole. Every node is taken when neither nodeset nor tags is given"),
		option.Query("nodeset", "Nodes by name: a nodeset like cpn-u25-[01-40] picks the nodes it names, and other text picks the names holding it", param.Example("nodeset", "cpn-u25")),
		option.Query("tags", "Nodes with any of these tags, as tags select nodes elsewhere in the API", param.Example("tags", "rack=u25")),
		option.QueryBool("bmc", "Put the ranges holding BMC ports first when true, or the other ports when false. Without it the ranges are in address order"),
	)
	fuego.Patch(nodes, "/tags/{action}", h.NodeTags, about("Add and remove tags", "Nodes"),
		option.Description("Update nodes tags by nodeset and/or tags"),
		option.Path("action", "option to add or remove tags", param.Example("action", "add | remove")),
		filterNodes,
	)
	fuego.Get(nodes, "/token/{interface}", h.NodeBootToken, about("Create boot tokens", "Provisioning"),
		option.Description("Create a boot token for the provision server. Used for debugging requests made by images"),
		option.Path("interface", "interface token will be created for", param.Example("interface", "boot | bmc")),
		filterNodes,
		option.Query("arch", "Arch to embed in the token, used in templates when the node has no arch set. Simulates the arch detected during a PXE boot", param.Example("arch", "x86_64")),
	)
	fuego.Get(nodes, "/render/{template}", h.NodeRender, about("Render provision templates", "Provisioning"),
		option.Description("Render a template for each node as the provision server would send it, without recording a provision stage. Templates are read from disk on each request, so an edit is seen at once and one that does not parse is an error. The output holds a working boot token and whatever secrets the template renders, so the nodes have to be named: a request with neither nodeset nor tags is refused"),
		option.Path("template", "What to render: ipxe, kickstart, user_data, meta_data, butane, ignition (the butane template translated), proxmox, a key of the image's provision_templates, or a template file name", param.Example("template", "ipxe | kickstart | idrac-config.json.tmpl")),
		filterNodes,
		option.Query("interface", "Interface the node is taken to boot from", param.Default("boot"), param.Example("interface", "boot | bmc")),
		renderOptions,
	)
	fuego.Get(nodes, "/check", h.NodeCheck, about("Check boot files and templates", "Provisioning"),
		option.Description("Check that each node can be served its boot: the kernel and initrds of its image are readable at the paths rendered for the node, with their .sig files when the image is verified, and the iPXE script and every template in the image's provision_templates render. Files are read on the machine the API runs on. Every node is taken when neither nodeset nor tags is given"),
		filterNodes,
		renderOptions,
	)
	fuego.Patch(nodes, "/image", h.NodeBootImage, about("Change a node's boot image", "Nodes"),
		option.Description("Update nodes boot image by nodeset and/or tags"),
		filterNodes,
	)

	fuego.Post(images, "", h.BootImageAdd, about("Add and edit images", "Boot images"), option.Description("Add images"))
	fuego.Get(images, "", h.BootImageList, about("List images", "Boot images"), option.Description("List all images"))
	fuego.Delete(images, "", h.BootImageDelete, about("Delete images", "Boot images"), option.Description("Delete images by name"), filterNames)
	fuego.Get(images, "/find", h.BootImageFind, about("Find images", "Boot images"), option.Description("Find images by name"), filterNames)

	fuego.Post(users, "", h.UserStore, about("Add users", "Users"), option.Description("Add new user"))
	fuego.Get(users, "", h.UserList, about("List users", "Users"), option.Description("List all users"), option.Query("usernames", "Filter by usernames", param.Example("username", "admin,user")))
	fuego.Delete(users, "/{usernames}", h.UserDelete, about("Delete users", "Users"),
		option.Description("Delete users"),
		usernamesExample,
	)
	fuego.Patch(users, "/{usernames}/role", h.UserRole, about("Change a user's role", "Users"),
		option.Description("Update users role"),
		usernamesExample,
	)
	fuego.Patch(users, "/{usernames}/enable", h.UserEnable, about("Enable and disable users", "Users"),
		option.Description("Update users enable"),
		usernamesExample,
	)

	fuego.Post(auth, "/signin", h.AuthSignin, about("Sign in", "Account"),
		option.Description("signin user"),
		option.Security(openapi3.NewSecurityRequirement()),
	)
	fuego.Post(auth, "/signup", h.AuthSignup, about("Sign up", "Account"),
		option.Description("Signup user"),
		option.Security(openapi3.NewSecurityRequirement()),
	)
	fuego.Delete(auth, "/signout", h.AuthSignout, about("Sign out", "Account"),
		option.Description("Signout user"),
		option.Security(openapi3.NewSecurityRequirement()),
	)
	fuego.Post(auth, "/token", h.AuthToken, about("Create API tokens", "Account"),
		option.Description("Create API token"),
		option.Middleware(h.authMiddleware),
	)
	fuego.Patch(auth, "/reset", h.AuthReset, about("Change own password", "Account"),
		option.Description("Change password"),
		option.Middleware(h.authMiddleware),
	)

	fuego.Post(db, "/restore", h.Restore, about("Restore a database backup", "Server"), option.Description("Restore a backup of the DB"))
	fuego.Get(db, "/dump", h.Dump, about("Download a database backup", "Server"), option.Description("Get a backup of the DB"))

	fuego.Get(bmc, "/nics", h.BmcNICs, about("List network ports", "BMC"),
		option.Description("List the host network ports each node's BMC reports"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/status", h.BmcStatus, about("See system status", "BMC"),
		option.Description("Get redfish system information from node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/power", h.BmcPower, about("Power nodes and restart BMCs", "BMC"),
		option.Description("Change the power state of node(s), or restart their BMCs"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/indicator", h.BmcIndicator, about("Turn the locate light on or off", "BMC"),
		option.Description("Turn the location indicator of node(s) on or off"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/asset-tag", h.BmcAssetTag, about("Set asset tags", "BMC"),
		option.Description("Set the asset tag of node(s), from a template or a table of serial numbers"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/location", h.BmcLocation, about("Set rack locations", "BMC"),
		option.Description("Set the rack location of node(s), rendered per node from templates"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/identity/defaults", h.BmcIdentityDefaults, about("See asset tag and location defaults", "BMC"),
		option.Description("Get the location and asset tag templates configured as defaults"),
	)
	// a websocket, which OpenAPI cannot describe, so it stays out of the spec
	fuego.GetStd(bmc, "/console", h.BmcConsole, option.Hide())
	fuego.Get(bmc, "/bios", h.BmcBios, about("Read BIOS attributes", "BIOS"),
		option.Description("Get the BIOS attributes of node(s), with any staged changes"),
		option.Query("names", "Keep only these attributes", param.Example("names", "SriovGlobalEnable,ProcCStates")),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/bios", h.BmcSetBios, about("Stage BIOS changes", "BIOS"),
		option.Description("Stage BIOS attribute changes on node(s), applied at the given apply time"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/bios/discard", h.BmcDiscardBios, about("Discard staged BIOS changes", "BIOS", model.PermissionDelete),
		option.Description("Discard the BIOS changes staged on node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/bios/registry", h.BmcBiosRegistry, about("Read the BIOS attribute registry", "BIOS"),
		option.Description("Get the attribute registry describing one node's BIOS attributes"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/secureboot", h.BmcSecureBoot, about("See secure boot state and keys", "Secure boot"),
		option.Description("Get the UEFI secure boot state of node(s), with any staged change, and with keys the certificates and signatures in their databases"),
		option.QueryBool("keys", "Also list the certificates and signatures in each database", param.Default(false)),
		option.Query("databases", "With keys, keep only these databases", param.Example("databases", "db,dbx")),
		option.QueryBool("defaults", "With keys, also list the factory default databases, such as dbDefault", param.Default(false)),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/secureboot", h.BmcSetSecureBoot, about("Turn secure boot on or off", "Secure boot"),
		option.Description("Turn UEFI secure boot on or off on node(s), from their next boot"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/secureboot/keys", h.BmcAddSecureBootKey, about("Enroll secure boot keys", "Secure boot"),
		option.Description("Enroll certificates or a signature in a secure boot database on node(s). A key already enrolled is left alone"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/secureboot/keys/delete", h.BmcDeleteSecureBootKey, about("Delete secure boot keys", "Secure boot", model.PermissionDelete),
		option.Description("Remove a certificate or signature from a secure boot database on node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/secureboot/reset", h.BmcResetSecureBootKeys, about("Reset secure boot keys", "Secure boot", model.PermissionDelete),
		option.Description("Reset the secure boot databases of node(s) to their factory keys, or delete their keys"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/jobs", h.BmcJobs, about("See the job queue", "BMC"),
		option.Description("Get the redfish job queue of node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/jobs/clear", h.BmcClearJobs, about("Clear jobs", "BMC", model.PermissionDelete),
		option.Description("Clear redfish jobs from node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/sel", h.BmcSEL, about("Read the system event log", "BMC"),
		option.Description("Get the system event log of node(s)"),
		option.QueryInt("limit", "Keep only the newest entries of each node, this many. 0 returns the whole log", param.Default(defaultSELLimit)),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/sel/clear", h.BmcClearSEL, about("Clear the system event log", "BMC", model.PermissionDelete),
		option.Description("Clear the system event log on node(s)"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/configure/auto", h.BmcAutoConfigure, about("Set iDRAC auto configure", "BMC"),
		option.Description("Set iDRAC(s) to auto configure once, on the node's next boot. Dell only"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/configure/import", h.BmcImportConfig, about("Import an iDRAC config", "BMC"),
		option.Description("Have iDRAC(s) download a config template from the provision server and import it. Dell only"),
		filterNodes,
		selectAll,
	)
	fuego.Get(bmc, "/firmware/updates", h.BmcFirmwareUpdates, about("List available firmware updates", "Firmware"),
		option.Description("List the firmware updates iDRAC(s) found in their last catalog comparison, which firmware/install with apply_update false runs. Dell only"),
		filterNodes,
		selectAll,
	)
	fuego.Post(bmc, "/firmware/install", h.BmcInstallFirmware, about("Check and install firmware", "Firmware"),
		option.Description("Compare the firmware of iDRAC(s) against an update catalog, and with apply_update queue the updates. A comparison waits for its job to finish, up to bmc.node_timeout. Dell only"),
		filterNodes,
		selectAll,
	)

	fuego.Get(sw, "/scan", h.SwitchScan, about("Scan switch ports and MAC tables", "Switches"),
		option.Description("Read the MAC table and interface status of switch(es), merged per port, with the node that already holds each MAC"),
		filterNodes,
		selectAll,
	)
	fuego.Get(sw, "/{nodeset}/lldp", h.SwitchGetLLDP, about("Read LLDP neighbours", "Switches"),
		option.Description("Get switch LLDP info"),
		option.Query("ports", "Filter by port name", param.Example("ports", "Et1,Et2")),
		// filterNodes,
	)

	fuego.Get(roles, "", h.GetRoles, about("List roles and permissions", "Roles"),
		option.Description("Get roles and permissions"),
		option.Query("name", "Filter by name", param.Example("name", "admin,user")),
	)
	fuego.Patch(roles, "", h.PatchRoles, about("Edit a role's permissions", "Roles"),
		option.Description("Edit role permissions"),
	)
	fuego.Post(roles, "", h.PostRoles, about("Add roles", "Roles"),
		option.Description("Add roles"),
	)
	fuego.Delete(roles, "/{names}", h.DeleteRoles, about("Delete roles", "Roles"),
		option.Description("Delete roles"),
		option.Path("names", "Delete by name", param.Example("names", "admin,user")),
	)

	fuego.Get(config, "", h.GetSettings, about("List settings", "Settings"),
		option.Description("Every setting with its help text, its value and where that value comes from. Secrets never carry a value"),
	)
	fuego.Patch(config, "", h.PatchSettings, about("Change settings", "Settings"),
		option.Description("Store settings in the database or remove them from it. Nothing is saved unless every change is valid"),
	)

	h.routes = describeRoutes(s.OpenAPI.Description())
	// the console is hidden from the spec, so it is described here
	h.routes[model.Permission{Method: http.MethodGet, Path: "/v1/bmc/console"}] = routeInfo{Summary: "Open the serial console", Group: "BMC", Kind: model.PermissionChange}
}

// kindExtension carries a route's permission kind in the spec when its HTTP method does not say it
const kindExtension = "x-grendel-kind"

// about names a route for the role editor and the API docs: a short summary, the group it is listed under, and its kind when the method would give the wrong one, such as a POST that deletes
func about(summary, group string, kind ...model.PermissionKind) func(*fuego.BaseRoute) {
	return func(r *fuego.BaseRoute) {
		option.Summary(summary)(r)
		option.Tags(group)(r)
		if len(kind) > 0 {
			if r.Operation.Extensions == nil {
				r.Operation.Extensions = map[string]any{}
			}
			r.Operation.Extensions[kindExtension] = string(kind[0])
		}
	}
}
