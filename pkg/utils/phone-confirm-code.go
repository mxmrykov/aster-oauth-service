package utils

import "math/rand"

func GetConfirmCode() int {
	return 100_000 + rand.Intn(899_999)
}
