package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/intob/ddb/ctl"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go func(ctx context.Context, cancel context.CancelFunc) {
		err := ctl.Start(ctx)
		if err != nil {
			fmt.Printf("failed to start ctl server: %s\r\n", err)
			cancel()
		}
	}(ctx, cancel)

	go func(cancel context.CancelFunc) {
		switch <-sigs {
		case syscall.SIGINT:
			fmt.Println("\r\nreceived SIGINT")
		case syscall.SIGTERM:
			fmt.Println("\r\nreceived SIGTERM")
		}
		cancel()
	}(cancel)

	wg.Add(1)
	go transport.Listen(ctx, wg)

	subscribeToPingAndAck()
	subscribeToAck()

	wg.Wait()
	fmt.Println("all routines ended")
}

func subscribeToPingAndAck() {
	rcvEvents := make(chan *event.Event)
	event.Subscribe(&event.Sub{
		MatchFunc: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC && e.Rpc.Type == rpc.TYPE_PING
		},
		Rcvr: rcvEvents,
	})
	go func(rcvEvents <-chan *event.Event) {
		for e := range rcvEvents {
			fmt.Println("I got pinged", e.Rpc.Id)
			transport.SendRpc(&transport.AddrRpc{
				Rpc: &rpc.Rpc{
					Id:   e.Rpc.Id,
					Type: rpc.TYPE_ACK,
				},
				Addr: e.Addr,
			})
		}
	}(rcvEvents)
}

func subscribeToAck() {
	rcvEvents := make(chan *event.Event)
	event.Subscribe(&event.Sub{
		MatchFunc: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC && e.Rpc.Type == rpc.TYPE_ACK
		},
		Rcvr: rcvEvents,
		Once: true,
	})
	go func(rcvEvents <-chan *event.Event) {
		for e := range rcvEvents {
			fmt.Println("my ping was acknowledged", e.Rpc.Id)
		}
	}(rcvEvents)
}
