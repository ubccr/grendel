// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package gormstore

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"reflect"

	"gorm.io/gorm/schema"
)

// MAC serializer is used to avoid improper []byte storage when using the rqlite driver
// IP Prefix serializer is used to avoid type wrapping / implementing value / scan on netip.Prefix

// MACSerializer net.HardwareAddr serializer
type MACSerializer net.HardwareAddr

// Scan implements serializer interface
func (MACSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	switch value := dbValue.(type) {
	case []byte:
		return field.Set(ctx, dst, value)
	case string:
		mac, _ := net.ParseMAC(value)
		return field.Set(ctx, dst, mac)
	default:
		return fmt.Errorf("unsupported data %#v", dbValue)
	}
}

// Value implements serializer interface
func (MACSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	var str string
	switch v := fieldValue.(type) {
	case net.HardwareAddr:
		str = v.String()
	default:
		return nil, fmt.Errorf("incorrect input type, received: %T, wanted: net.HardwareAddr", fieldValue)
	}

	return str, nil
}

// MACSerializer net.HardwareAddr serializer
type IPPrefixSerializer netip.Prefix

// Scan implements serializer interface
func (IPPrefixSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	var str string

	switch value := dbValue.(type) {
	case []byte:
		str = string(value)
	case string:
		str = value
	default:
		return fmt.Errorf("unsupported data %#v", dbValue)
	}
	ip, err := netip.ParsePrefix(str)
	if err != nil {
		return err
	}
	return field.Set(ctx, dst, ip)
}

// Value implements serializer interface
func (IPPrefixSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	var str string
	switch v := fieldValue.(type) {
	case netip.Prefix:
		str = v.String()
	default:
		return nil, fmt.Errorf("incorrect input type, received: %T, wanted: netip.Prefix", fieldValue)
	}

	return str, nil
}
