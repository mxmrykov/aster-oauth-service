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
	Eaid          string `json:"EAID"`
	Signature     string `json:"signature"`
	SignatureDate string `json:"signatureDate"`
	jwt.StandardClaims
}

type SidToken struct {
	Iaid          string `json:"IAID"`
	Asid          string `json:"ASID"`
	SignatureDate string `json:"signatureDate"`
	jwt.StandardClaims
}
