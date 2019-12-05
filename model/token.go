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
	ExpiresAt int64               `json:"exp"`
	Firmware  firmware.BootLoader `json:"fw"`
}

type BootClaims struct {
	MAC      string              `json:"mac"`
	BootSpec string              `json:"bootspec"`
	Firmware firmware.BootLoader `json:"firmware"`
	jwt.StandardClaims
}

func init() {
	viper.SetDefault("jwt_tokens", true)
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

func NewBootToken(mac string, bootspec string, fwtype firmware.BootLoader) (string, error) {
	claims := &BootClaims{
		MAC:      mac,
		BootSpec: bootspec,
		Firmware: fwtype,
	}
	claims.ExpiresAt = time.Now().Add(time.Second * 60 * 60).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(viper.GetString("secret")))
	if err != nil {
		return "", err
	}

	return t, nil
}

func NewFirmwareToken(mac string, fwtype firmware.BootLoader) (string, error) {
	if !viper.GetBool("jwt_tokens") {
		return fmt.Sprintf("%s/%d", mac, fwtype), nil
	}

	claims := &FirmwareClaims{
		Firmware:  fwtype,
		ExpiresAt: time.Now().Add(time.Second * 60).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(viper.GetString("secret")))
	if err != nil {
		return "", err
	}

	return t, nil
}

func ParseFirmwareToken(tokenString string) (firmware.BootLoader, error) {
	if !viper.GetBool("jwt_tokens") {
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

		return firmware.BootLoader(i), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &FirmwareClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(viper.GetString("secret")), nil
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
