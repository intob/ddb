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
	ev, _ := event.Subscribe(func(e *event.Event) bool {
		return e.Topic == event.Rpc &&
			e.Rpc.Type == rpc.ListAddr
	})
	for e := range ev {
		fmt.Println("rcvd list addr rpc", e.Rpc.Id)
		resp := &rpc.ListAddrBody{
			AddrList: make([]string, contact.Count()-1), // exclude contact's own address
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
				Type: rpc.Ack,
				Body: respBytes,
			},
			Addr: e.Addr,
		})
		if err != nil {
			fmt.Println("failed to send list addr ACK rpc:", err)
		}
	}
}
