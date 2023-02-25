package rpcin

import (
	"fmt"

	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func SubscribeToPingAndAck() {
	ev, _ := event.Subscribe(event.RpcTypeFilter(rpc.Ping))
	for e := range ev {
		detail, _ := e.Detail.(event.RpcDetail)
		err := transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   detail.Rpc.Id,
				Type: rpc.Ack,
			},
			Addr: detail.Addr,
		})
		if err != nil {
			fmt.Println("failed to send ping ack rpc:", err)
		}
	}
}
