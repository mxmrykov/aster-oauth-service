package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
)

func NewAccessRefreshToken(i model.AccessRefreshToken, signature string, access ...bool) (string, error) {
	exp := time.Now().Add(time.Hour * 24 * 30).Unix()
	if access != nil && access[0] {
		exp = time.Now().Add(time.Minute * 15).Unix()
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, model.AccessRefreshToken{
		Iaid:          i.Iaid,
		Eaid:          i.Eaid,
		SignatureDate: time.Now().Format(time.RFC3339),
		Signature:     i.Signature,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp,
			Issuer:    "proxy.aster.oauth",
		},
	})
	return unsignedToken.SignedString([]byte(signature))
}

func ValidateXAuthToken(XAuthToken, signature string) (model.XAuthToken, error) {
	parsedToken, err := jwt.ParseWithClaims(
		XAuthToken,
		&model.XAuthToken{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(signature), nil
		},
	)

	if claims, ok := parsedToken.Claims.(*model.XAuthToken); ok && parsedToken.Valid {
		return *claims, nil
	}

	return model.XAuthToken{}, err
}

func ValidateAccessRefreshToken(XAuthToken, signature string) (model.AccessRefreshToken, error) {
	parsedToken, err := jwt.ParseWithClaims(
		XAuthToken,
		&model.AccessRefreshToken{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(signature), nil
		},
	)

	if claims, ok := parsedToken.Claims.(*model.AccessRefreshToken); ok && parsedToken.Valid {
		return *claims, nil
	}

	return model.AccessRefreshToken{}, err
}

func ValidateAsidToken(token, signature string) (model.SidToken, error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&model.SidToken{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(signature), nil
		},
	)

	if claims, ok := parsedToken.Claims.(*model.SidToken); ok && parsedToken.Valid {
		return *claims, nil
	}

	return model.SidToken{}, err
}
