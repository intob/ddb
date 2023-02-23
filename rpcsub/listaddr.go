package rpcsub

import (
	"context"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func SubscribeToListAddrRpc(ctx context.Context) {
	rcv := make(chan *event.Event)
	_, err := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC &&
				e.Rpc.Type == rpc.TYPE_LIST_ADDR
		},
		Rcvr: rcv,
	})
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to list addr rpc: %w", err))
	}
	for e := range rcv {
		fmt.Println("rcvd list addr rpc", e.Rpc.Id)
		resp := &rpc.ListAddrBody{
			AddrList: make([]string, contact.Count()-1),
		}
		i := 0
		for _, c := range contact.GetAll() {
			// don't send contact's own address
			if c.Addr.String() == e.Addr.String() {
				continue
			}
			resp.AddrList[i] = c.Addr.String()
			i++
		}
		respBytes, err := cbor.Marshal(resp)
		if err != nil {
			fmt.Println("failed to marshal list addr resp:", err)
		}
		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   e.Rpc.Id,
				Type: rpc.TYPE_ACK,
				Body: respBytes,
			},
			Addr: e.Addr,
		})
		if err != nil {
			fmt.Println("failed to send list addr ACK rpc:", err)
		}
	}
}
