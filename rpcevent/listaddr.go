package rpcevent

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func SubscribeToListAddrRpc() {
	ev, _ := event.Subscribe(event.RpcTypeFilter(rpc.ListAddr))
	for e := range ev {
		detail, _ := e.Detail.(event.RpcDetail)
		fmt.Println("rcvd list addr rpc", detail.Rpc.Id)
		resp := &rpc.ListAddrBody{
			AddrList: make([]string, contact.Count()-1), // exclude contact's own address
		}
		i := 0
		for _, c := range contact.GetAll() {
			// don't send contact's own address
			if c.Addr.String() == detail.Addr.String() {
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
				Id:   detail.Rpc.Id,
				Type: rpc.Ack,
				Body: respBytes,
			},
			Addr: detail.Addr,
		})
		if err != nil {
			fmt.Println("failed to send list addr ack rpc:", err)
		}
	}
}
