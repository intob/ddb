package rpcout

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func SendListAddrRpcToNewContacts(ctx context.Context) {
	ev, _ := event.Subscribe(func(e *event.Event) bool {
		if e.Topic != event.ContactAdded {
			return false
		}
		_, ok := e.Detail.(event.ContactAddedDetail)
		return ok
	})
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-ev:
			detail, _ := e.Detail.(event.ContactAddedDetail)
			ListAddr(ctx, detail.Addr)
		}
	}
}

func ListAddr(ctx context.Context, addr *net.UDPAddr) {
	rpcId := rpc.RandId()
	ev, _ := event.SubscribeOnce(event.RpcIdFilter(rpcId))
	transport.SendRpc(&transport.AddrRpc{
		Rpc: &rpc.Rpc{
			Id:   rpcId,
			Type: rpc.ListAddr,
		},
		Addr: addr,
	})
	timeout := time.NewTimer(time.Second)
	select {
	case r := <-ev:
		handleListAddrResp(r)
	case <-timeout.C:
		fmt.Println("list addr rpc timed out")
		break
	case <-ctx.Done():
		break
	}
}

func handleListAddrResp(e *event.Event) error {
	detail, ok := e.Detail.(event.RpcDetail)
	if !ok {
		return fmt.Errorf("unexpected type %T in event detail", e.Detail)
	}
	listAddr := &rpc.ListAddrBody{}
	err := cbor.Unmarshal(detail.Rpc.Body, listAddr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal list addr resp body: %w", err)
	}
	fmt.Println("addresses:", listAddr)
	for _, a := range listAddr.AddrList {
		udpAddr, err := net.ResolveUDPAddr("udp", a)
		if err != nil {
			fmt.Println("failed to resolve udp addr:", err)
			continue
		}
		contact.Put(&contact.Contact{
			Addr: udpAddr,
		})
	}
	return nil
}
