package rpcsub

import (
	"context"
	"fmt"

	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func SubscribeToPingAndAck(ctx context.Context) {
	ev, _ := event.Subscribe(func(e *event.Event) bool {
		return e.Topic == event.Rpc && e.Rpc.Type == rpc.Ping
	})
	for e := range ev {
		err := transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   e.Rpc.Id,
				Type: rpc.Ack,
			},
			Addr: e.Addr,
		})
		if err != nil {
			fmt.Println("failed to send ping ack rpc:", err)
		}
	}
}
