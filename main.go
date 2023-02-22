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
	"github.com/intob/ddb/gossip"
	"github.com/intob/ddb/healthcheck"
	"github.com/intob/ddb/listaddr"
	"github.com/intob/ddb/rpcsub"
	"github.com/intob/ddb/transport"
)

func init() {
	contact.Seed(time.Now().UnixNano())
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

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

	wg.Add(1)
	go transport.Listen(ctx, wg)

	wg.Add(1)
	go rpcsub.SubscribeToPingAndAck(ctx, wg)

	wg.Add(1)
	go rpcsub.SubscribeToGetRpc(ctx, wg)

	wg.Add(1)
	go rpcsub.SubscribeToStoreRpc(ctx, wg)

	wg.Add(1)
	go rpcsub.SubscribeToListAddrRpc(ctx, wg)

	wg.Add(1)
	go gossip.PropagateStoreRpcs(ctx, wg)

	wg.Add(1)
	go healthcheck.PingContacts(ctx, wg)

	wg.Add(1)
	go listaddr.ListAddrOfNewContacts(ctx, wg)

	for _, c := range contact.GetAll() {
		go listaddr.ListAddr(c.Addr)
	}

	go func() {
		<-ctx.Done()
		event.UnsubscribeAll()
	}()

	fmt.Println("all routines started")
	wg.Wait()
	fmt.Println("all routines ended")
}
