package listaddr

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

func ListAddrOfNewContacts(ctx context.Context, wg *sync.WaitGroup) {
	defer fmt.Println("ListAddrOfNewContacts done")
	defer wg.Done()
	rcv := make(chan *event.Event)
	event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_CONTACT_ADDED
		},
		Rcvr: rcv,
	})
	for {
		select {
		case <-ctx.Done():
			return
		case contactAdded := <-rcv:
			ListAddr(contactAdded.Addr)
		}
	}
}

func ListAddr(addr *net.UDPAddr) {
	rpcId, err := rpc.RandId()
	if err != nil {
		fmt.Println("failed to get rand rpc id:", err)
		return
	}
	resp := make(chan *event.Event)
	event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC && bytes.Equal(*e.Rpc.Id, *rpcId)
		},
		Rcvr: resp,
		Once: true,
	})
	transport.SendRpc(&transport.AddrRpc{
		Rpc: &rpc.Rpc{
			Id:   rpcId,
			Type: rpc.TYPE_LIST_ADDR,
		},
		Addr: addr,
	})
	timeout := time.NewTimer(time.Second)
	select {
	case r := <-resp:
		handleListAddrResp(r)
	case <-timeout.C:
		fmt.Println("list addr rpc timed out")
	}
}

func handleListAddrResp(e *event.Event) error {
	listAddr := &rpc.ListAddrBody{}
	err := cbor.Unmarshal(e.Rpc.Body, listAddr)
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
