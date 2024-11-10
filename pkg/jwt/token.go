package jwt

import (
	"github.com/golang-jwt/jwt"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
)

func NewAccessRefreshToken(t string) (string, error) {
	return "", nil
}

func ValidateXAuthToken(XAuthToken, signature string) (model.XAuthToken, error) {
	parsedToken, err := jwt.ParseWithClaims(
		XAuthToken,
		&model.XAuthToken{},
		func(token *jwt.Token) (interface{}, error) {
			return signature, nil
		},
	)

	if claims, ok := parsedToken.Claims.(*model.XAuthToken); ok && parsedToken.Valid {
		return *claims, nil
	}

	return model.XAuthToken{}, err
}
