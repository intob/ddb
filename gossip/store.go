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

type KV struct {
	Key   string
	Value []byte
}

func SubscribeToStoreRpc(ctx context.Context, wg *sync.WaitGroup) {
	rcvEvents := make(chan *event.Event)
	subId, err := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC &&
				e.Rpc.Type == rpc.TYPE_STORE
		},
		Rcvr: rcvEvents,
	})
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to store rpc: %w", err))
	}
	go func(rcvEvents <-chan *event.Event) {
		for e := range rcvEvents {
			fmt.Println("rcvd store rpc")
			kv := &KV{}
			err := cbor.Unmarshal(e.Rpc.Body, kv)
			if err != nil {
				fmt.Println("failed to unmarshal store rpc body:", err)
				continue
			}
			store.Set(kv.Key, &kv.Value)
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
		fmt.Println("SubscribeToStoreRpc done")
		wg.Done()
	}(rcvEvents)
	<-ctx.Done()
	event.Unsubscribe(subId)
}
