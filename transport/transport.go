package transport

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
)

const BUFFER_SIZE = 1024

var (
	rpcOut = make(chan *AddrRpc)
	laddr  *net.UDPAddr
)

type AddrRpc struct {
	Rpc  *rpc.Rpc
	Addr *net.UDPAddr
}

func init() {
	// parse args
	for i, arg := range os.Args {
		if arg == "--listen" && len(os.Args) > i+1 {
			l, err := net.ResolveUDPAddr("udp", os.Args[i+1])
			if err != nil {
				fmt.Println("failed to resolve listen address:", err)
				os.Exit(1)
			}
			laddr = l
		}
	}

	// fallback to any available port
	if laddr == nil {
		var err error
		laddr, err = net.ResolveUDPAddr("udp", ":0")
		if err != nil {
			panic(err)
		}
	}
}

func SendRpc(addrRpc *AddrRpc) error {
	if addrRpc.Addr == nil {
		return fmt.Errorf("rpc addr must not be nil")
	}
	if addrRpc.Rpc.Id == nil {
		return fmt.Errorf("rpc must have an id")
	}
	rpcOut <- addrRpc
	return nil
}

func Listen(ctx context.Context, wg *sync.WaitGroup) {
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		panic(fmt.Errorf("failed to listen udp: %w", err))
	}
	defer conn.Close()

	fmt.Println("listening on", conn.LocalAddr())

	for {
		select {
		case <-ctx.Done():
			fmt.Println("listener closed")
			wg.Done()
			return

		case r := <-rpcOut:
			b, err := rpc.PackRpc(r.Rpc)
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = conn.WriteToUDP(b, r.Addr)
			if err != nil {
				fmt.Println("failed to write rpc:", err)
			}

		default:
			buf := make([]byte, BUFFER_SIZE)
			conn.SetReadDeadline(time.Now().Add(time.Millisecond))
			n, raddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			r, err := rpc.UnpackRpc(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}
			addOrUpdateContact(raddr)
			event.Publish(&event.Event{
				Topic: event.TOPIC_RPC,
				Rpc:   r,
				Addr:  raddr,
			})
		}
	}
}

func addOrUpdateContact(raddr *net.UDPAddr) {
	c := contact.Get(raddr.String())
	if c == nil {
		// TODO: check if contact should be added
		contact.Put(&contact.Contact{
			Addr:     raddr,
			LastSeen: time.Now(),
		})
		return
	}
	c.LastSeen = time.Now()
}
