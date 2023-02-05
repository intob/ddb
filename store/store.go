package store

var store map[string][]byte

type KV struct {
	Key   string
	Value []byte
}

func init() {
	store = make(map[string][]byte)
}

func Get(key string) []byte {
	return store[key]
}

func Set(key string, value []byte) {
	store[key] = value
}

func Rm(key string) {
	delete(store, key)
}
