package model

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type FirmwareClaims struct {
	ExpiresAt int64    `json:"exp"`
	Firmware  Firmware `json:"fw"`
}

type BootClaims struct {
	MAC          string       `json:"mac"`
	BootSpec     string       `json:"bootspec"`
	Firmware     Firmware     `json:"firmware"`
	Architecture Architecture `json:"arch"`
	jwt.StandardClaims
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

func NewBootToken(mac string, bootspec string, fwtype Firmware, arch Architecture) (string, error) {
	claims := &BootClaims{
		MAC:          mac,
		BootSpec:     bootspec,
		Firmware:     fwtype,
		Architecture: arch,
	}
	claims.ExpiresAt = time.Now().Add(time.Second * 60).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//TODO fixme
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return t, nil
}

func NewFirmwareToken(fwtype Firmware) (string, error) {
	claims := &FirmwareClaims{
		Firmware:  fwtype,
		ExpiresAt: time.Now().Add(time.Second * 60).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//TODO fixme
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return t, nil
}

func ParseFirmwareToken(tokenString string) (Firmware, error) {
	token, err := jwt.ParseWithClaims(tokenString, &FirmwareClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		//TODO fixme
		return []byte("secret"), nil
	})

	if claims, ok := token.Claims.(*FirmwareClaims); ok && token.Valid {
		return claims.Firmware, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return 0, fmt.Errorf("Token is malformed")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return 0, fmt.Errorf("Token is expired")
		} else {
			return 0, fmt.Errorf("Error parsing token: ", err)
		}
	}

	return 0, fmt.Errorf("Failed to parsetoken: ", err)
}
