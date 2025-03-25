// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/pkg/model"
)

const (
	signingKey = "secret-signing-key"
)

func TestJwtClaims(t *testing.T) {
	assert := assert.New(t)

	claims := jwt.MapClaims{
		TokenUsername: "test-user",
		TokenRole:     model.RoleDisabled.String(),
	}
	token, err := NewToken(claims, signingKey)
	assert.Nil(err)

	tokenClaims, err := ParseToken(token, signingKey)
	assert.Nil(err)

	assert.Equal(tokenClaims.username, "test-user")
	assert.Equal(tokenClaims.role, "disabled")
}

func TestJwtExpire(t *testing.T) {
	assert := assert.New(t)

	claims := jwt.MapClaims{
		TokenUsername: "test-user",
		TokenRole:     model.RoleDisabled.String(),
		TokenExpire:   time.Now().Add(time.Second).Unix(),
	}
	token, err := NewToken(claims, signingKey)
	assert.Nil(err)

	_, err = ParseToken(token, signingKey)
	assert.Nil(err)

	time.Sleep(time.Second)

	_, err = ParseToken(token, signingKey)
	assert.ErrorIs(err, jwt.ErrTokenExpired)
}

func TestJwtChangeSecret(t *testing.T) {
	assert := assert.New(t)

	claims := jwt.MapClaims{
		TokenUsername: "test-user",
		TokenRole:     model.RoleDisabled.String(),
	}
	token, err := NewToken(claims, signingKey)
	assert.Nil(err)

	_, err = ParseToken(token, "different-key")
	assert.ErrorIs(err, jwt.ErrSignatureInvalid)
}

func TestJwtInvalidAlg(t *testing.T) {
	assert := assert.New(t)

	noneToken := "eyJhbGciOiJub25lIn0.eyJ1c2VybmFtZSI6InRlc3QiLCJyb2xlIjoiZGlzYWJsZWQifQ."
	_, err := ParseToken(noneToken, signingKey)
	assert.ErrorIs(err, jwt.ErrTokenSignatureInvalid)
}
