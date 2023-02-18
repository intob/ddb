package contact

import (
	"fmt"
	"os"
	"sync"
)

type Contact struct {
	Addr string
}

var (
	contacts = make([]*Contact, 0)
	mutex    = &sync.Mutex{}
)

func init() {
	for i, arg := range os.Args {
		if arg == "--contact" && len(os.Args) > i+1 {
			Put(&Contact{
				Addr: os.Args[i+1],
			})
			fmt.Println("added contact", os.Args[i+1])
		}
	}
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
