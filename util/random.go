package util

import (
	"math/rand"
	"strings"
)

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
	number   = "0123456789"
)

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomNumber(n int) string {
	var sb strings.Builder
	k := len(number)

	for i := 0; i < n; i++ {
		c := number[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomBookingCode() string {
	return RandomString(8)
}

func RandomResetPasswordToken() string {
	return RandomNumber(5)
}
