package api

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenUsername = "username"
	TokenRole     = "role"
	TokenExpire   = "exp"
)

type claims struct {
	username string
	role     string
}

func NewToken(claims jwt.MapClaims, signingKey string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(signingKey))
}

func ParseToken(tokenString, signingKey string) (*claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))

	token, err := parser.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	rawClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse raw token claims")
	}

	username, ok := rawClaims[TokenUsername].(string)
	if !ok || username == "" {
		return nil, errors.New("failed to parse token claims")
	}
	role, ok := rawClaims[TokenRole].(string)
	if !ok || role == "" {
		return nil, errors.New("failed to parse token claims")
	}

	return &claims{
		username: username,
		role:     role,
	}, nil
}
