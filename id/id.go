package id

import (
	"crypto/rand"
	"encoding/hex"
)

type Id []byte

func Rand(len int) (*Id, error) {
	b := make(Id, len)
	_, err := rand.Read(b)
	return &b, err
}

func (id *Id) String() string {
	return hex.EncodeToString(*id)
}
