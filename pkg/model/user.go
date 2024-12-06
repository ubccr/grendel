// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"errors"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

// Enum for Role
type Role int

const (
	RoleAdmin Role = iota + 1
	RoleUser
	RoleDisabled
)

var ErrInvalidRole = errors.New("Invalid role")

// Return the string of a Role
func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleUser:
		return "user"
	case RoleDisabled:
		return "disabled"
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
	case "disabled":
		return RoleDisabled, nil
	default:
		return RoleUser, ErrInvalidRole
	}
}
