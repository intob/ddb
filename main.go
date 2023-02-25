package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/ctl"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/gossip"
	"github.com/intob/ddb/rpcin"
	"github.com/intob/ddb/rpcout"
	"github.com/intob/ddb/transport"
)

func init() {
	contact.Seed(time.Now().UnixNano())
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context, cancel context.CancelFunc) {
		err := ctl.Start()
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

	go transport.Listen(ctx)
	go gossip.PropagateStoreRpcs(ctx)
	go rpcout.PingContacts(ctx)
	go rpcout.SendListAddrRpcToNewContacts(ctx)
	go rpcin.SubscribeToPingAndAck()
	go rpcin.SubscribeToGetRpc()
	go rpcin.SubscribeToStoreRpc()
	go rpcin.SubscribeToListAddrRpc()

	for _, c := range contact.GetAll() {
		go rpcout.ListAddr(ctx, c.Addr)
	}

	go func() {
		<-ctx.Done()
		event.UnsubscribeAll()
	}()

	fmt.Println("all routines started")
	<-ctx.Done()
	fmt.Println("all routines ended")
}
