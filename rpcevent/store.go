package rpcevent

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

func SubscribeToStoreRpc() {
	ev, _ := event.Subscribe(event.RpcTypeFilter(rpc.Store))
	for e := range ev {
		// type assertion already checked in filter
		detail, _ := e.Detail.(event.RpcDetail)
		fmt.Println("rcvd store rpc")
		b := &rpc.StoreBody{}
		err := cbor.Unmarshal(detail.Rpc.Body, b)
		if err != nil {
			fmt.Println("failed to unmarshal store rpc body:", err)
			continue
		}
		store.Set(b.Key, &b.Value, b.Modified)
		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   detail.Rpc.Id,
				Type: rpc.Ack,
			},
			Addr: detail.Addr,
		})
		if err != nil {
			fmt.Println("failed to send store ack rpc:", err)
		}
	}
}
