package contact

import (
	"fmt"
	"net"
	"testing"
)

func init() {
	Seed(1)
}

func putContact(addr string) error {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	Put(&Contact{
		Addr: a,
	})
	return nil
}

func TestGetNonExistent(t *testing.T) {
	contacts = make(map[string]*Contact)
	if Get("lala") != nil {
		t.FailNow()
	}
}

func TestGetExisting(t *testing.T) {
	contacts = make(map[string]*Contact)
	addr, _ := net.ResolveUDPAddr("udp", "localhost:1992")
	Put(&Contact{
		Addr: addr,
	})
	if Get(addr.String()) == nil {
		t.FailNow()
	}
}

func TestRandWhenEmpty(t *testing.T) {
	contacts = make(map[string]*Contact)
	_, err := Rand(make([]string, 0))
	if err == nil {
		t.Fatalf("should have returned an error")
	}
}

func TestRandAllExcluded(t *testing.T) {
	contacts = make(map[string]*Contact)
	excluded := make([]string, 0)
	for i := 10; i < 99; i++ {
		addrStr := fmt.Sprintf("localhost:10%v", i)
		addr, _ := net.ResolveUDPAddr("udp", addrStr)
		Put(&Contact{
			Addr: addr,
		})
		excluded = append(excluded, addr.String())
	}
	c, err := Rand(excluded)
	if c != nil || err == nil {
		t.Fatalf("all contacts should have been excluded")
	}
}

func TestRandAllButOneExcluded(t *testing.T) {
	contacts = make(map[string]*Contact)
	excluded := make([]string, 0)
	for i := 10; i < 99; i++ {
		addrStr := fmt.Sprintf("localhost:10%v", i)
		putContact(addrStr)
		if i != 50 {
			excluded = append(excluded, addrStr)
		}
	}
	c, err := Rand(excluded)
	if err != nil {
		t.Fatalf("error should be nil: %s", err)
	}
	if c.Addr.Port != 1050 {
		t.Fatalf("contact's port should have been 1050, was %v", c.Addr.Port)
	}
}

func TestRandomizedList(t *testing.T) {
	for i := 10; i < 100; i++ {
		putContact(fmt.Sprintf("localhost:10%v", i))
	}
	initial := make([]*Contact, Count())
	i := 0
	for _, c := range GetAll() {
		initial[i] = c
		i++
	}
	randomized := RandomizedList()
	countDontMatch := 0
	for i := 0; i < len(initial); i++ {
		if initial[i].Addr.String() != randomized[i].Addr.String() {
			countDontMatch++
		}
	}
	if countDontMatch < len(initial)/4 {
		t.FailNow()
	}
}
