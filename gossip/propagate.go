package gossip

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

// number of nodes to propagate rpcs to
const r = 2

type LogEntry struct {
	Time time.Time
}

func PropagateStoreRpcs(ctx context.Context, wg *sync.WaitGroup) {
	log := make(map[string]*LogEntry, 0)
	rcvEvents := make(chan *event.Event)
	subId, err := event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC &&
				e.Rpc.Type == rpc.TYPE_STORE
		},
		Rcvr: rcvEvents,
	})
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to rpcs: %w", err))
	}
	go func() {
		for e := range rcvEvents {
			if log[e.Rpc.Id.String()] != nil {
				fmt.Println("already seen, won't propagate")
				continue
			}
			log[e.Rpc.Id.String()] = &LogEntry{time.Now()}
			// pick r contacts at random, other than the sender
			// TODO: exclude all previous senders also to limit useless gossip
			contacts := make([]*contact.Contact, 0)
			exclude := []string{e.Addr.String()}
			for {
				l := len(contacts)
				if l == r || l == contact.Count()-1 {
					break
				}
				// pick random contact
				randContact, err := contact.Rand(exclude)
				if err != nil {
					fmt.Println("failed to pick random contact:", err)
					break
				}
				contacts = append(contacts, randContact)
				exclude = append(exclude, randContact.Addr.String())
			}
			for _, c := range contacts {
				err := transport.SendRpc(&transport.AddrRpc{
					Rpc:  e.Rpc,
					Addr: c.Addr,
				})
				if err != nil {
					fmt.Println("failed to propagate rpc:", err)
				}
				fmt.Println("propagated rpc to", c.Addr.String())
			}
		}
		fmt.Println("PropagateStoreRpcs done")
		wg.Done()
	}()
	<-ctx.Done()
	event.Unsubscribe(subId)
}
