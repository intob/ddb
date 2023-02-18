package store

import (
	"fmt"
	"sync"

	"github.com/intob/ddb/id"
)

const (
	keyByteLen = 16
)

var (
	store = make(map[string][]byte)
	mutex = &sync.Mutex{}
)

func Get(key string) []byte {
	return store[key]
}

func Set(key string, value *[]byte) {
	mutex.Lock()
	store[key] = *value
	fmt.Println("will store:", string(*value))
	mutex.Unlock()
}

func Rm(key string) {
	mutex.Lock()
	delete(store, key)
	mutex.Unlock()
}

func RandKey() (string, error) {
	key, err := id.Rand(keyByteLen)
	return key.String(), err
}
