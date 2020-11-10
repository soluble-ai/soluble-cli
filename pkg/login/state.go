package login

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
)

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func MakeState() string {
	const n = 64
	var bb bytes.Buffer
	bb.Grow(n)
	l := uint16(len(letters))
	b := make([]byte, 2)
	for i := 0; i < n; i++ {
		_, _ = rand.Read(b)
		bb.WriteRune(letters[binary.BigEndian.Uint16(b)%l])
	}
	return bb.String()
}
