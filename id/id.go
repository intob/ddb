package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Id []byte

func Rand(len int) *Id {
	b := make(Id, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Errorf("failed to generate random id: %w", err))
	}
	return &b
}

func (id *Id) String() string {
	return hex.EncodeToString(*id)
}
