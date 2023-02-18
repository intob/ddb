package id

import "crypto/rand"

func RandId(len int) ([]byte, error) {
	b := make([]byte, len)
	_, err := rand.Read(b)
	return b, err
}
