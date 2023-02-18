package store

import "sync"

var (
	store = make(map[string][]byte)
	mutex = &sync.Mutex{}
)

type KV struct {
	Key   string
	Value []byte
}

func Get(key string) []byte {
	return store[key]
}

func Set(key string, value []byte) {
	mutex.Lock()
	store[key] = value
	mutex.Unlock()
}

func Rm(key string) {
	mutex.Lock()
	delete(store, key)
	mutex.Unlock()
}
