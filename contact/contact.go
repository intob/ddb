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
	Addr      *net.UDPAddr
	LastSeen  time.Time
	RoundTrip time.Duration
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
	addr := c.Addr.String()
	if contacts[addr] != nil {
		return
	}
	mutex.Lock()
	contacts[addr] = c
	mutex.Unlock()
}

func Get(addr string) *Contact {
	return contacts[addr]
}

func GetAll() []Contact {
	list := make([]Contact, len(contacts))
	i := 0
	for _, c := range contacts {
		list[i] = *c
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

func Rand(exclude map[string]bool) (*Contact, error) {
	if len(contacts) == 0 {
		return nil, fmt.Errorf("no contacts")
	}
	if len(exclude) >= len(contacts) {
		return nil, fmt.Errorf("all contacts excluded")
	}
	var chosen *Contact
	for chosen == nil {
		n := rnd.Intn(len(contacts) - 1)
		i := 0
		for _, c := range contacts {
			if i == n {
				if !exclude[c.Addr.String()] {
					chosen = c
					break
				}
			}
			i++
		}
	}
	return chosen, nil
}

func UpdateRoundTrip(addr string, d time.Duration) {
	mutex.Lock()
	c := contacts[addr]
	if c != nil {
		c.RoundTrip = d
	}
	mutex.Unlock()
}
