package model

import (
	"github.com/golang-jwt/jwt"
)

type XAuthToken struct {
	Asid          string `json:"ASID"`
	SignatureDate string `json:"signatureDate"`
	jwt.StandardClaims
}

type AccessRefreshToken struct {
	Iaid          string `json:"IAID"`
	SignatureDate string `json:"signatureDate"`
	jwt.StandardClaims
}
