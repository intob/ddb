package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/rpc"
)

const BUFFER_SIZE_BYTES = 1024

var laddr *net.UDPAddr

type AddrRpc struct {
	Rpc  *rpc.Rpc
	Addr string
}

func init() {
	// parse args
	for i, arg := range os.Args {
		if arg == "-l" && len(os.Args) > i+1 {
			l, err := net.ResolveUDPAddr("udp", os.Args[i+1])
			if err != nil {
				panic(err)
			}
			laddr = l
		}
		if arg == "-c" && len(os.Args) > i+1 {
			contact.Put(&contact.Contact{
				Addr: os.Args[i+1],
			})
			fmt.Println("added contact", os.Args[i+1])
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

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go func(cancel context.CancelFunc) {
		switch <-sigs {
		case syscall.SIGINT:
			fmt.Println("\r\nreceived SIGINT")
		case syscall.SIGTERM:
			fmt.Println("\r\nreceived SIGTERM")
		}
		cancel()
	}(cancel)

	rpcIn := make(chan *AddrRpc)
	rpcOut := make(chan *AddrRpc)

	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, rpcIn chan<- *AddrRpc, rpcOut <-chan *AddrRpc) {
		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		fmt.Println("listening on", conn.LocalAddr().String())

		for {
			select {
			case <-ctx.Done():
				fmt.Println("closing UDP connection...")
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
				fmt.Println("received: ", string(buf[:n]), " from ", raddr)
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
	}(ctx, wg, rpcIn, rpcOut)

	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, rpcIn <-chan *AddrRpc) {
		select {
		case <-ctx.Done():
			fmt.Println("handler exiting...")
			wg.Done()
			return
		case r := <-rpcIn:
			fmt.Println("got msg!", r.Rpc.Id)
			// send to eventbus
		}
	}(ctx, wg, rpcIn)

	wg.Wait()
	fmt.Println("all routines ended")
}
