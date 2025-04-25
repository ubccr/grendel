// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"hash"`
	Role         string    `json:"role"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type RoleViewList []RoleView

type RoleView struct {
	Name                     string         `json:"name"`
	PermissionList           PermissionList `json:"permission_list"`
	UnassignedPermissionList PermissionList `json:"unassigned_permission_list"`
}
type PermissionList []Permission

type Permission struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

func (r *RoleView) Scan(value interface{}) error {
	data, ok := value.(string)
	if !ok {
		return errors.New("incompatible type")
	}
	var role RoleView
	err := json.Unmarshal([]byte(data), &role)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	*r = role
	return nil
}

// Enum for Role
type Role int

const (
	RoleAdmin Role = iota + 1
	RoleUser
	RoleReadOnly
	roleCount
)

var ErrInvalidRole = errors.New("invalid role")

// Return the string of a Role
func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleUser:
		return "user"
	case RoleReadOnly:
		return "read-only"
	default:
		return "Unknown role"
	}
}

// New role from string
func RoleFromString(name string) (Role, error) {
	switch name {
	case "admin":
		return RoleAdmin, nil
	case "user":
		return RoleUser, nil
	case "read-only":
		return RoleReadOnly, nil
	default:
		return RoleReadOnly, ErrInvalidRole
	}
}

func AllRoles() []Role {
	roles := make([]Role, roleCount)
	for i := range roles {
		roles[i] = Role(i)
	}

	return roles
}
