package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/intob/ddb/ctl"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/id"
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
	go transport.StartListener(ctx, wg)

	wg.Add(1)
	go transport.StartHandler(ctx, wg)

	doTestStuff()

	wg.Wait()
	fmt.Println("all routines ended")
}

func doTestStuff() {
	rcvEvents := make(chan *event.Event)
	subId, _ := id.RandId(event.SUB_ID_BYTE_LEN)
	event.Subscribe(&event.Sub{
		Id: subId,
		MatchFunc: func(event *event.Event) bool {
			if event.Rpc.Type == rpc.TYPE_PING {
				return true
			}
			return false
		},
		Rcvr: rcvEvents,
	})
	go func(rcvEvents chan *event.Event) {
		for e := range rcvEvents {
			fmt.Println("got subscribed event!", hex.EncodeToString(e.Id))
		}
	}(rcvEvents)
}
