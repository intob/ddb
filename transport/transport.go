package transport

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
)

const BUFFER_SIZE_BYTES = 1024

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

func Laddr() string {
	return laddr.String()
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

func StartListener(ctx context.Context, wg *sync.WaitGroup) {
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("listening on", conn.LocalAddr().String())

	for {
		select {
		case <-ctx.Done():
			fmt.Println("listener exiting...")
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
				fmt.Printf("failed to write rpc: %s\r\n", err)
			}

			fmt.Println("sent rpc to ", r.Addr)

		default:
			buf := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(time.Second))
			n, raddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			r, err := rpc.UnpackRpc(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}
			event.Publish(&event.Event{
				Topic: event.TOPIC_RPC,
				Rpc:   r,
				Raddr: raddr,
			})
		}
	}
}
