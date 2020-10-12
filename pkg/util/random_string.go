package util

import (
	"math/rand"
	"time"
)

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

var charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenRandomString creates and returns a random string of provided length
func GenRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
