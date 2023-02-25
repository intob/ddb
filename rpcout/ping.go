package rpcout

import (
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
	lastSeenThreshold = time.Second // will not ping if last seen recently
	pingTimeout       = time.Second
)

func PingContacts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			timeout(ctx, lastSeenThreshold)
			for _, c := range contact.GetAll() {
				// skip if seen recently
				if c.LastSeen.After(time.Now().Add(-lastSeenThreshold)) {
					continue
				}

				rid := rpc.RandId()

				addrStr := c.Addr.String()

				ev, _ := event.SubscribeOnce(event.RpcIdFilter(rid))

				transport.SendRpc(&transport.AddrRpc{
					Rpc: &rpc.Rpc{
						Id:   rid,
						Type: rpc.Ping,
					},
					Addr: c.Addr,
				})
				timeStart := time.Now()

				pingTimer := time.NewTimer(pingTimeout)
				select {
				case <-ctx.Done():
					return
				case <-pingTimer.C:
					contact.Rm(addrStr)
					fmt.Println("ping timed out, removed contact", addrStr)
				case <-ev:
					timeEnd := time.Now()
					contact.UpdateRoundTrip(c.Addr.String(), timeEnd.Sub(timeStart))
				}
				randWait := rand.Intn(int(lastSeenThreshold))
				timeout(ctx, time.Duration(randWait))
			}
		}
	}
}

func timeout(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
	}
}
