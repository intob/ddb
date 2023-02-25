package store

import (
	"sync"
	"time"
)

var (
	store = make(map[string]*Entry)
	mutex = &sync.Mutex{}
)

type Entry struct {
	Value    []byte
	Modified time.Time
}

func Get(key string) *Entry {
	return store[key]
}

func Set(key string, value []byte, modified time.Time) {
	mutex.Lock()
	defer mutex.Unlock()
	e := store[key]
	if e == nil {
		store[key] = &Entry{
			Value:    value,
			Modified: modified,
		}
		return
	}
	store[key].Value = value
	store[key].Modified = modified
}

func Rm(key string) {
	mutex.Lock()
	delete(store, key)
	mutex.Unlock()
}
