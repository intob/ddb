package healthcheck

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

const (
	roundPeriod       = time.Second      // time between each new round
	contactPeriod     = time.Second      // 1 contact per contact period
	lastSeenThreshold = 10 * time.Second // will not ping if last seen recently
	pingTimeout       = time.Second
)

func PingContacts(ctx context.Context, wg *sync.WaitGroup) {
	defer fmt.Println("PingContacts done")
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			list := contact.RandomizedList()
			for _, c := range list {
				// skip if seen recently
				if c.LastSeen.After(time.Now().Add(-lastSeenThreshold)) {
					continue
				}

				rid, err := rpc.RandId()
				if err != nil {
					fmt.Println("failed to generate random id:", err)
				}

				rcv := make(chan *event.Event)
				event.Subscribe(&event.Sub{
					Filter: func(e *event.Event) bool {
						return e.Topic == event.TOPIC_RPC &&
							bytes.Equal(*e.Rpc.Id, *rid)
					},
					Rcvr: rcv,
					Once: true,
				})

				transport.SendRpc(&transport.AddrRpc{
					Rpc: &rpc.Rpc{
						Id:   rid,
						Type: rpc.TYPE_PING,
					},
					Addr: c.Addr,
				})

				timeout := time.NewTimer(pingTimeout)
				select {
				case <-ctx.Done():
					return
				case <-rcv:
					fmt.Println("got ping ACK")
					time.Sleep(contactPeriod)
				case <-timeout.C:
					contact.Rm(c.Addr.String())
					fmt.Println("ping timed out, removed contact")
				}
			}
			roundTimeout := time.NewTimer(roundPeriod)
			select {
			case <-ctx.Done():
				return
			case <-roundTimeout.C:
			}
		}
	}
}
