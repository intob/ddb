package healthcheck

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

const (
	roundPeriod       = time.Second      // time between each new round
	contactPeriod     = time.Millisecond // time between each contact in the round
	lastSeenThreshold = 10 * time.Second // will not ping if last seen recently
	pingTimeout       = time.Second
)

func PingContacts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for _, c := range contact.GetAll() {
				// skip if seen recently
				if c.LastSeen.After(time.Now().Add(-lastSeenThreshold)) {
					continue
				}

				rid, err := rpc.RandId()
				if err != nil {
					fmt.Println("failed to generate random id:", err)
				}

				ev, _ := event.SubscribeOnce(func(e *event.Event) bool {
					return e.Topic == event.TOPIC_RPC &&
						bytes.Equal(*e.Rpc.Id, *rid)
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
				case <-timeout.C:
					contact.Rm(c.Addr.String())
					fmt.Println("ping timed out, removed contact", c.Addr.String())
				case <-ev:
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
