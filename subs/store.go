package subs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

type StoreRpcBody struct {
	Key      string
	Value    []byte
	Modified time.Time
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
	fmt.Println("SubscribeToStoreRpc subId:", subId)
	go func() {
		for e := range rcvEvents {
			fmt.Println("rcvd store rpc")
			b := &StoreRpcBody{}
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
		fmt.Println("SubscribeToStoreRpc done")
		wg.Done()
	}()
	<-ctx.Done()
	event.Unsubscribe(subId)
}
