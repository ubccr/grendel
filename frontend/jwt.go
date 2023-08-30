package frontend

import (
	"crypto"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"time"
)

type SigningMethodHMAC struct {
	Name string
	Hash crypto.Hash
}

type AuthClaims struct {
	User string `json:"user"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func Sign(user, role string) (string, time.Time, error) {
	mySigningKey := []byte(viper.GetString("frontend.signingKey"))
	ExpiresAt := time.Now().Add(8 * time.Hour)
	httpExpires := ExpiresAt.UTC()

	claims := AuthClaims{
		user,
		role,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "grendel",
			Subject:   "auth",
			//ID:        "1",
			//Audience:  []string{"somebody_else"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", time.Now(), err
	}

	return ss, httpExpires, nil
}

func Verify(token string) (string, string, error) {

	t, err := jwt.ParseWithClaims(token, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("frontend.signingKey")), nil
	})

	if claims, ok := t.Claims.(*AuthClaims); ok && t.Valid {
		return claims.User, claims.Role, nil
	}
	return "", "", err
}
