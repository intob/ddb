package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Id string

func Rand(len int) Id {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Errorf("failed to generate random id: %w", err))
	}
	return Id(hex.EncodeToString(b))
}
