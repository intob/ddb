package contact

import (
	"fmt"
	"math/rand"
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
	rnd      *rand.Rand
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

func Seed(seed int64) {
	rnd = rand.New(rand.NewSource(seed))
}

func Put(c *Contact) {
	if contacts[c.Addr.String()] != nil {
		return
	}
	mutex.Lock()
	contacts[c.Addr.String()] = c
	mutex.Unlock()
}

func Get(addr string) *Contact {
	return contacts[addr]
}

func GetAll() []*Contact {
	list := make([]*Contact, len(contacts))
	i := 0
	for _, c := range contacts {
		list[i] = c
		i++
	}
	return list
}

func Rm(addr string) {
	mutex.Lock()
	delete(contacts, addr)
	mutex.Unlock()
}

func Count() int {
	return len(contacts)
}

func Rand(exclude []string) (*Contact, error) {
	if len(exclude) >= len(contacts) {
		return nil, fmt.Errorf("all contacts excluded")
	}
	var chosen *Contact
	for chosen == nil {
		n := rnd.Intn(len(contacts) - 1)
		i := 0
		for _, c := range contacts {
			if i == n {
				if !isExcluded(c.Addr.String(), exclude) {
					chosen = c
					break
				}
			}
			i++
		}
	}
	return chosen, nil
}

func isExcluded(addr string, exclude []string) bool {
	for _, ex := range exclude {
		if ex == addr {
			return true
		}
	}
	return false
}
