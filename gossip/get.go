package gossip

import (
	"context"
	"fmt"
	"sync"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

func SubscribeToGetRpc(ctx context.Context, wg *sync.WaitGroup) {
	rcvEvents := make(chan *event.Event)
	subId, err := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC &&
				e.Rpc.Type == rpc.TYPE_GET
		},
		Rcvr: rcvEvents,
	})
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to store rpc: %w", err))
	}
	go func(rcvEvents <-chan *event.Event) {
		for e := range rcvEvents {
			fmt.Println("rcvd get rpc", e.Rpc.Id)
			entry := store.Get(string(e.Rpc.Body))
			var entryBytes []byte
			if entry != nil {
				entryBytes, err = cbor.Marshal(entry)
				if err != nil {
					fmt.Println("failed to marshal entry")
					continue
				}
			}
			err := transport.SendRpc(&transport.AddrRpc{
				Rpc: &rpc.Rpc{
					Id:   e.Rpc.Id,
					Type: rpc.TYPE_ACK,
					Body: entryBytes,
				},
				Addr: e.Addr,
			})
			if err != nil {
				fmt.Println("failed to send get ACK rpc:", err)
			}
		}
		wg.Done()
	}(rcvEvents)
	<-ctx.Done()
	event.Unsubscribe(subId)
}
