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
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/eventstore"
)

type Handler struct {
	DB     store.Store
	Events eventstore.Store
}

func NewHandler(db store.Store) (*Handler, error) {
	h := &Handler{
		DB: db,
	}

	return h, nil
}

func (h *Handler) SetupRoutes(s *fuego.Server) {

	// enable frontend if api is listening on tcp socket
	if viper.IsSet("api.listen") {
		fuego.Handle(s, "/ui/", setupFrontend())
		fuego.Handle(s, "/{$}", http.RedirectHandler("/ui", http.StatusMovedPermanently))
	}

	// Examples
	nsExample := param.Example("nodeset", "cpn-i10-[04-05],cpn-h22-33")
	usernamesExample := option.Path("usernames", "target usernames", param.Example("usernames", "user1,user2"))

	// Params
	filterNodes := fuego.GroupOptions(option.Query("nodeset", "Filter by nodeset. Minimum of one query parameter is required", nsExample), option.Query("tags", "Filter by tags. Minimum of one query parameter is required", param.Example("tags", "a01,ib,test")))
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

	// Routes
	fuego.Get(grendel, "/events", h.GetEvents)

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
	fuego.Get(users, "", h.UserList, option.Description("List all users"))
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

	fuego.Get(bmc, "", h.BmcQuery,
		option.Description("Get redfish info from node(s)"),
		filterNodes,
	)
	fuego.Post(bmc, "/power/os", h.BmcOsPower,
		option.Description("Change power status of node(s)"),
		filterNodes,
	)
	fuego.Post(bmc, "/power/bmc", h.BmcPower,
		option.Description("Reboot node(s) BMC"),
		filterNodes,
	)
	fuego.Delete(bmc, "/sel", h.BmcSelClear,
		option.Description("Clear system event log on node(s)"),
		filterNodes,
	)
	fuego.Get(bmc, "/jobs", h.BmcJobList,
		option.Description("Get redfish jobs from node(s)"),
		filterNodes,
	)
	fuego.Delete(bmc, "/jobs/{jids}", h.BmcJobDelete,
		option.Description("Delete redfish jobs from node(s) by JID"),
		option.Path("jids", "Redfish Job IDs. Use 'JID_CLEARALL' to clear all jobs",
			param.Example("jids", "JID_000000000001,JID_000000000002"),
		),
		filterNodes,
	)
	fuego.Post(bmc, "/configure/auto", h.BmcAutoConfigure,
		option.Description("Set BMC to autoconfigure"),
		filterNodes,
	)
	fuego.Post(bmc, "/configure/import", h.BmcImportConfiguration,
		option.Description("Manually import system configuration to BMC"),
		filterNodes,
	)
	fuego.Get(bmc, "/metrics", h.BmcMetricReports,
		option.Description("Get metric reports by nodeset"),
		filterNodes,
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
