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

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/firmware"
)

type FirmwareClaims struct {
	ExpiresAt int64          `json:"exp"`
	Firmware  firmware.Build `json:"fw"`
}

type BootClaims struct {
	ID  string `json:"id"`
	MAC string `json:"mac"`
	jwt.StandardClaims
}

func init() {
	viper.SetDefault("provision.jwt_tokens", true)
}

func (p *FirmwareClaims) Valid() error {
	now := time.Now().Unix()
	if now <= p.ExpiresAt {
		return nil
	}

	vErr := new(jwt.ValidationError)
	delta := time.Unix(now, 0).Sub(time.Unix(p.ExpiresAt, 0))
	vErr.Inner = fmt.Errorf("token is expired by %v", delta)
	vErr.Errors = jwt.ValidationErrorExpired

	return vErr
}

func NewBootToken(id, mac string) (string, error) {
	claims := &BootClaims{
		ID:  id,
		MAC: mac,
	}
	claims.ExpiresAt = time.Now().Add(time.Second * 60 * 60).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(viper.GetString("provision.secret")))
	if err != nil {
		return "", err
	}

	return t, nil
}

func NewFirmwareToken(mac string, fwtype firmware.Build) (string, error) {
	if !viper.GetBool("provision.jwt_tokens") {
		return fmt.Sprintf("%s/%d", mac, fwtype), nil
	}

	claims := &FirmwareClaims{
		Firmware:  fwtype,
		ExpiresAt: time.Now().Add(time.Second * 60 * 60).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(viper.GetString("provision.secret")))
	if err != nil {
		return "", err
	}

	return t, nil
}

func ParseFirmwareToken(tokenString string) (firmware.Build, error) {
	if !viper.GetBool("provision.jwt_tokens") {
		pathElements := strings.Split(tokenString, "/")
		if len(pathElements) != 2 {
			return 0, fmt.Errorf("Invalid tftp file path: %s", tokenString)
		}

		_, err := net.ParseMAC(pathElements[0])
		if err != nil {
			return 0, fmt.Errorf("Invalid MAC address: %s", pathElements[0])
		}

		i, err := strconv.Atoi(pathElements[1])
		if err != nil {
			return 0, fmt.Errorf("Invalid firmware type: %s", pathElements[1])
		}

		return firmware.Build(i), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &FirmwareClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(viper.GetString("provision.secret")), nil
	})

	if claims, ok := token.Claims.(*FirmwareClaims); ok && token.Valid {
		return claims.Firmware, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return 0, fmt.Errorf("Token is malformed")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return 0, fmt.Errorf("Token is expired")
		} else {
			return 0, fmt.Errorf("Error parsing token: %s", err)
		}
	}

	return 0, fmt.Errorf("Failed to parsetoken: %s", err)
}
