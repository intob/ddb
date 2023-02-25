package rpcin

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

func SubscribeToGetRpc() {
	ev, _ := event.Subscribe(event.RpcTypeFilter(rpc.Get))
	for e := range ev {
		detail, _ := e.Detail.(event.RpcDetail)
		entry := store.Get(string(detail.Rpc.Body))
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
				Id:   detail.Rpc.Id,
				Type: rpc.Ack,
				Body: entryBytes,
			},
			Addr: detail.Addr,
		})
		if err != nil {
			fmt.Println("failed to send get ack rpc:", err)
		}
	}
}
