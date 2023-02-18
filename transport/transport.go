package transport

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/intob/ddb/rpc"
)

const BUFFER_SIZE_BYTES = 1024

var (
	rpcIn  chan *AddrRpc
	rpcOut chan *AddrRpc
	laddr  *net.UDPAddr
)

type AddrRpc struct {
	Rpc  *rpc.Rpc
	Addr string
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

	rpcIn = make(chan *AddrRpc)
	rpcOut = make(chan *AddrRpc)
}

func Laddr() string {
	return laddr.String()
}

func SendRpc(addrRpc *AddrRpc) {
	rpcOut <- addrRpc
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
			addr, err := net.ResolveUDPAddr("udp", r.Addr)
			if err != nil {
				fmt.Printf("failed to resolve udp addr: %s\r\n", err)
			}
			_, err = conn.WriteToUDP(b, addr)
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
			rpcIn <- &AddrRpc{
				Rpc:  r,
				Addr: raddr.String(),
			}
		}
	}
}

func StartHandler(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("handler exiting...")
			wg.Done()
			return
		case r := <-rpcIn:
			fmt.Println("got msg!", r.Rpc.Type)
			// send to eventbus
		}
	}
}
