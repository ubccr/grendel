// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"encoding/json"

	"github.com/hako/branca"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/firmware"
	"github.com/ubccr/grendel/internal/util"
)

type BootClaims struct {
	ID  string `json:"id"`
	MAC string `json:"mac"`
}

func init() {
	viper.SetDefault("provision.token_ttl", 60*60)

	if !viper.IsSet("provision.secret") {
		secret, err := util.GenerateSecret(16)
		if err != nil {
			panic(err)
		}
		viper.SetDefault("provision.secret", secret)
	}
}

func NewBootToken(id, mac string) (string, error) {
	claims := &BootClaims{
		ID:  id,
		MAC: mac,
	}

	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	b := branca.NewBranca(viper.GetString("provision.secret"))
	b.SetTTL(viper.GetUint32("provision.token_ttl"))

	token, err := b.EncodeToString(string(jsonBytes))
	if err != nil {
		return "", err
	}

	return token, nil
}

func ParseBootToken(token string) (*BootClaims, error) {
	b := branca.NewBranca(viper.GetString("provision.secret"))
	b.SetTTL(viper.GetUint32("provision.token_ttl"))

	message, err := b.DecodeToString(token)
	if err != nil {
		return nil, err
	}

	var claims BootClaims
	err = json.Unmarshal([]byte(message), &claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

func NewFirmwareToken(mac string, fwtype firmware.Build) (string, error) {
	b := branca.NewBranca(viper.GetString("provision.secret"))
	b.SetTTL(viper.GetUint32("provision.token_ttl"))

	token, err := b.EncodeToString(fwtype.String())
	if err != nil {
		return "", err
	}

	return token, nil
}

func ParseFirmwareToken(token string) (firmware.Build, error) {
	b := branca.NewBranca(viper.GetString("provision.secret"))
	b.SetTTL(viper.GetUint32("provision.token_ttl"))

	message, err := b.DecodeToString(token)
	if err != nil {
		return 0, err
	}

	return firmware.NewFromString(message), nil
}
