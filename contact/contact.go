package contact

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type Contact struct {
	Addr     *net.UDPAddr
	LastSeen time.Time
}

var (
	contacts = make(map[string]*Contact, 0)
	mutex    = &sync.Mutex{}
)

func init() {
	for i, arg := range os.Args {
		if arg == "--contact" && len(os.Args) > i+1 {
			addr, err := net.ResolveUDPAddr("udp", os.Args[i+1])
			if err != nil {
				fmt.Println("failed to resolve udp addr:", err)
				continue
			}
			Put(&Contact{
				Addr: addr,
			})
			fmt.Println("added contact", os.Args[i+1])
		}
	}
}

func Put(c *Contact) {
	mutex.Lock()
	defer mutex.Unlock()
	contacts[c.Addr.String()] = c
}

func Get(addr string) *Contact {
	return contacts[addr]
}

func Rm(addr string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(contacts, addr)
}

func Count() int {
	return len(contacts)
}
