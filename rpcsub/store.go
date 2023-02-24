package rpcsub

import (
	"context"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

func SubscribeToStoreRpc(ctx context.Context) {
	ev, _ := event.Subscribe(func(e *event.Event) bool {
		return e.Topic == event.TOPIC_RPC &&
			e.Rpc.Type == rpc.TYPE_STORE
	})
	for e := range ev {
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
