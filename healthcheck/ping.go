package healthcheck

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

const (
	contactPeriod     = time.Millisecond // time between each contact in the round
	lastSeenThreshold = time.Second      // will not ping if last seen recently
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
					return e.Topic == event.Rpc &&
						bytes.Equal(*e.Rpc.Id, *rid)
				})

				transport.SendRpc(&transport.AddrRpc{
					Rpc: &rpc.Rpc{
						Id:   rid,
						Type: rpc.Ping,
					},
					Addr: c.Addr,
				})
				timeStart := time.Now()

				timeout := time.NewTimer(pingTimeout)
				select {
				case <-ctx.Done():
					return
				case <-timeout.C:
					contact.Rm(c.Addr.String())
					fmt.Println("ping timed out, removed contact", c.Addr.String())
				case <-ev:
					timeEnd := time.Now()
					c.RoundTrip = timeEnd.Sub(timeStart)
					fmt.Println(c.Addr.Port, c.RoundTrip.Microseconds())
				}
			}
			randTimeout := rand.Intn(int(lastSeenThreshold))
			roundTimeout := time.NewTimer(lastSeenThreshold + time.Duration(randTimeout))
			select {
			case <-ctx.Done():
				return
			case <-roundTimeout.C:
			}
		}
	}
}
