package rpcsub

import (
	"context"
	"fmt"
	"sync"

	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func SubscribeToPingAndAck(ctx context.Context, wg *sync.WaitGroup) {
	defer fmt.Println("SubscribeToPingAndAck done")
	defer wg.Done()
	rcvEvents := make(chan *event.Event)
	_, err := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC && e.Rpc.Type == rpc.TYPE_PING
		},
		Rcvr: rcvEvents,
	})
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to ping rpc: %w", err))
	}
	for e := range rcvEvents {
		err := transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   e.Rpc.Id,
				Type: rpc.TYPE_ACK,
			},
			Addr: e.Addr,
		})
		if err != nil {
			fmt.Println("failed to send ping ack rpc:", err)
		}
	}

}
