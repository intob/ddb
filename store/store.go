package store

import (
	"sync"

	"github.com/intob/ddb/id"
)

const (
	keyByteLen = 16
)

var (
	store = make(map[*id.Id][]byte)
	mutex = &sync.Mutex{}
)

func Get(key *id.Id) []byte {
	return store[key]
}

func Set(key *id.Id, value []byte) {
	mutex.Lock()
	store[key] = value
	mutex.Unlock()
}

func Rm(key *id.Id) {
	mutex.Lock()
	delete(store, key)
	mutex.Unlock()
}

func RandKey() (*id.Id, error) {
	return id.Rand(keyByteLen)
}
