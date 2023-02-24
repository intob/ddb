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

func SubscribeToGetRpc(ctx context.Context) {
	ev, _ := event.Subscribe(func(e *event.Event) bool {
		return e.Topic == event.Rpc &&
			e.Rpc.Type == rpc.Get
	})
	for e := range ev {
		fmt.Println("rcvd get rpc", e.Rpc.Id)
		entry := store.Get(string(e.Rpc.Body))
		var entryBytes []byte
		if entry != nil {
			var err error
			entryBytes, err = cbor.Marshal(entry)
			if err != nil {
				fmt.Println("failed to marshal entry")
				continue
			}
		}
		err := transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   e.Rpc.Id,
				Type: rpc.Ack,
				Body: entryBytes,
			},
			Addr: e.Addr,
		})
		if err != nil {
			fmt.Println("failed to send get ACK rpc:", err)
		}
	}
}
