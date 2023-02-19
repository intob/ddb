package contact

import (
	"crypto/rand"
	"fmt"
	"math/big"
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

func Rand(exclude []string) (*Contact, error) {
	if !haveCandidate(exclude) {
		return nil, fmt.Errorf("all contacts are excluded")
	}
	var chosen *Contact
	maxIndex := big.NewInt(int64(len(contacts) - 1))
	for chosen == nil {
		nBig, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to read random number: %w", err)
		}
		n := int(nBig.Int64())
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
