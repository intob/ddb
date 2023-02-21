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
	mutex.Lock()
	defer mutex.Unlock()
	contacts[c.Addr.String()] = c
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
	defer mutex.Unlock()
	delete(contacts, addr)
}

func Count() int {
	return len(contacts)
}

func Rand(exclude []string) (*Contact, error) {
	if !haveCandidate(exclude) {
		return nil, fmt.Errorf("all contacts are excluded")
	}
	var chosen *Contact
	for chosen == nil {
		n := rnd.Intn(len(contacts) - 1)
		i := 0
		for _, c := range contacts {
			if n == i {
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

func haveCandidate(exclude []string) bool {
	for _, c := range contacts {
		isExcluded := false
		for _, ex := range exclude {
			if c.Addr.String() == ex {
				isExcluded = true
				break
			}
		}
		if !isExcluded {
			return true
		}
	}
	return false
}

func isExcluded(addr string, exclude []string) bool {
	for _, ex := range exclude {
		if ex == addr {
			return true
		}
	}
	return false
}

func RandomizedList() []*Contact {
	list := GetAll()
	rnd.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	return list
}
