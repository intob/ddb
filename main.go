package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/intob/ddb/contact"
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

	wg.Add(1)
	go subscribeToPingAndAck(ctx, wg)

	wg.Add(1)
	go subscribeToRpcFromContacts(ctx, wg)

	fmt.Println("all routines started")
	wg.Wait()
	fmt.Println("all routines ended")
}

func subscribeToPingAndAck(ctx context.Context, wg *sync.WaitGroup) {
	rcvEvents := make(chan *event.Event)
	subId, _ := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC && e.Rpc.Type == rpc.TYPE_PING
		},
		Rcvr: rcvEvents,
	})
	go func(rcvEvents <-chan *event.Event) {
		for e := range rcvEvents {
			fmt.Println("rcvd ping", e.Rpc.Id)
			transport.SendRpc(&transport.AddrRpc{
				Rpc: &rpc.Rpc{
					Id:   e.Rpc.Id,
					Type: rpc.TYPE_ACK,
				},
				Addr: e.Addr,
			})
		}
		fmt.Println("subscribeToPingAndAck done")
		wg.Done()
	}(rcvEvents)
	<-ctx.Done()
	event.Unsubscribe(subId)
}

func subscribeToRpcFromContacts(ctx context.Context, wg *sync.WaitGroup) {
	rpcEvents := make(chan *event.Event)
	subId, _ := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC && contact.Get(e.Addr.String()) != nil
		},
		Rcvr: rpcEvents,
	})
	go func(rpcEvents <-chan *event.Event, wg *sync.WaitGroup) {
		for e := range rpcEvents {
			fmt.Println("got rpc from contact", e.Rpc.Id)
			c := contact.Get(e.Addr.String())
			fmt.Println("contact", c.LastSeen.Format(time.RFC3339))
		}
		wg.Done()
	}(rpcEvents, wg)
	<-ctx.Done()
	event.Unsubscribe(subId)
	fmt.Println("subscribeToRpcFromContacts done")
}
