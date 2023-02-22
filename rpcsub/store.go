package rpcsub

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

func SubscribeToStoreRpc(ctx context.Context, wg *sync.WaitGroup) {
	defer fmt.Println("SubscribeToStoreRpc done")
	defer wg.Done()
	rcvEvents := make(chan *event.Event)
	_, err := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC &&
				e.Rpc.Type == rpc.TYPE_STORE
		},
		Rcvr: rcvEvents,
	})
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to store rpc: %w", err))
	}
	for e := range rcvEvents {
		fmt.Println("rcvd store rpc")
		b := &rpc.StoreBody{}
		err := cbor.Unmarshal(e.Rpc.Body, b)
		if err != nil {
			fmt.Println("failed to unmarshal store rpc body:", err)
			continue
		}
		store.Set(b.Key, &b.Value, b.Modified)
		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   e.Rpc.Id,
				Type: rpc.TYPE_ACK,
			},
			Addr: e.Addr,
		})
		if err != nil {
			fmt.Println("failed to send store ack rpc:", err)
		}
	}

}
