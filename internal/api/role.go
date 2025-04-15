package api

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/pkg/model"
)

type GetRolesResponse struct {
	Roles model.RoleViewList `json:"roles"`
}

type PostRolesRequest struct {
	Role          string `json:"role"`
	InheritedRole string `json:"inherited_role"`
}

type PatchRolesRequest struct {
	Role        string               `json:"role"`
	Permissions model.PermissionList `json:"permission_list"`
}

func (h *Handler) GetRoles(ctx fuego.ContextNoBody) (*GetRolesResponse, error) {
	filter := ctx.QueryParam("name")

	roles, err := h.DB.GetRoles()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get roles",
		}
	}
	var res GetRolesResponse

	if filter != "" {
		filterArr := strings.Split(filter, ",")
		for _, role := range roles {
			if slices.Contains(filterArr, role.Name) {
				res.Roles = append(res.Roles, role)
			}
		}
	} else {
		res.Roles = roles
	}

	if len(res.Roles) < 1 {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("failed to filter for role(s) filter=%s", filter),
			Title:  "Error",
			Detail: "failed to find role(s)",
		}
	}

	return &res, nil
}

func (h *Handler) PostRoles(ctx fuego.ContextWithBody[PostRolesRequest]) (*GenericResponse, error) {
	body, err := ctx.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get parse body",
		}
	}

	err = h.DB.AddRole(body.Role, body.InheritedRole)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to add role",
		}
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  fmt.Sprintf("Succesfully added role %s with permissions from %s", body.Role, body.InheritedRole),
		Changed: 1,
	}, nil
}

func (h *Handler) PatchRoles(ctx fuego.ContextWithBody[PatchRolesRequest]) (*GenericResponse, error) {
	body, err := ctx.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse body",
		}
	}

	for _, r := range model.AllRoles() {
		if body.Role != r.String() {
			continue
		}

		return nil, fuego.HTTPError{
			Err:    errors.New("default roles cannot be modified"),
			Title:  "Error",
			Detail: "Cannot modify default roles, please create a new role to modify",
		}
	}

	err = h.DB.UpdateRolePermissions(body.Role, body.Permissions)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "Failed to update permissions, ensure correct method and path names",
		}
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  fmt.Sprintf("succesfully updated permissions on role: %s", body.Role),
		Changed: 1,
	}, nil
}

func (h *Handler) DeleteRoles(ctx fuego.ContextNoBody) (*GenericResponse, error) {
	roleStr := ctx.PathParam("names")
	roles := strings.Split(roleStr, ",")

	for _, r := range model.AllRoles() {
		if !slices.Contains(roles, r.String()) {
			continue
		}

		return nil, fuego.HTTPError{
			Err:    errors.New("default roles cannot be modified"),
			Title:  "Error",
			Detail: "Cannot modify default roles, please create a new role to modify",
		}
	}

	err := h.DB.DeleteRole(roles)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "Failed to delete role, cannot remove roles which are still assigned to a user",
		}
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  "succesfully deleted role(s)",
		Changed: len(roles),
	}, nil
}
