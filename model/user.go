// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package model

import "time"

type User struct {
	Username     string    `json:"username"`
	PasswordHash []byte    `json:"hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

// Enum for Role
type Role int

const (
	RoleAdmin Role = iota + 1
	RoleDisabled
)

// Return the string of a Role
func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleDisabled:
		return "disabled"
	default:
		return "Unknown role"
	}
}
