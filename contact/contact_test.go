package contact

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	_, err := Rand(make(map[string]bool))
	if err == nil {
		t.Fatalf("should have returned an error")
	}
}

func TestRandAllExcluded(t *testing.T) {
	contacts = make(map[string]*Contact)
	excluded := make(map[string]bool)
	for i := 10; i < 99; i++ {
		addrStr := fmt.Sprintf("localhost:10%v", i)
		addr, _ := net.ResolveUDPAddr("udp", addrStr)
		Put(&Contact{
			Addr: addr,
		})
		excluded[addr.String()] = true
	}
	c, err := Rand(excluded)
	if c != nil || err == nil {
		t.Fatalf("all contacts should have been excluded")
	}
}

func TestRandAllButOneExcluded(t *testing.T) {
	contacts = make(map[string]*Contact)
	excluded := make(map[string]bool)
	for i := 10; i < 99; i++ {
		addrStr := fmt.Sprintf("127.0.0.1:10%v", i)
		putContact(addrStr)
		if i != 50 {
			excluded[addrStr] = true
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

func BenchmarkRand(b *testing.B) {
	contacts = make(map[string]*Contact)
	excluded := make(map[string]bool)
	for i := 10; i < 999; i++ {
		addrStr := fmt.Sprintf("127.0.0.1:10%v", i)
		putContact(addrStr)
		if i != 50 {
			excluded[addrStr] = true
		}
	}
	b.StartTimer()
	c, err := Rand(excluded)
	b.StopTimer()
	require.Nil(b, err)
	assert.Equal(b, 1050, c.Addr.Port)
}
