package contact

import (
	"sync"
)

type Contact struct {
	Addr string
}

var (
	contacts []*Contact
	mutex    = &sync.Mutex{}
)

func init() {
	contacts = make([]*Contact, 0)
}

func Put(c *Contact) {
	mutex.Lock()
	defer mutex.Unlock()
	contacts = append(contacts, c)
}

func Get(addr string) *Contact {
	for _, n := range contacts {
		if addr == n.Addr {
			return n
		}
	}
	return nil
}

func Rm(addr string) {
	mutex.Lock()
	defer mutex.Unlock()
	newList := make([]*Contact, 0)
	for _, n := range contacts {
		if addr != n.Addr {
			newList = append(newList, n)
		}
	}
	contacts = newList
}

func Count() int {
	return len(contacts)
}
