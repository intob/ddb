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
const (
	r                = 2
	logEntryLifetime = time.Second
	cleanLogPeriod   = time.Second
)

var (
	log   = make(map[string]*LogEntry, 0)
	mutex = &sync.Mutex{}
)

type LogEntry struct {
	Time time.Time
}

func PropagateStoreRpcs(ctx context.Context) {
	go cleanLog(ctx)
	events, _ := event.Subscribe(func(e *event.Event) bool {
		return e.Topic == event.TOPIC_RPC &&
			e.Rpc.Type == rpc.TYPE_STORE
	})
	for e := range events {
		rpcIdStr := e.Rpc.Id.String()
		mutex.Lock()
		if log[rpcIdStr] != nil {
			fmt.Println("already seen, won't propagate")
			mutex.Unlock()
			continue
		}
		log[rpcIdStr] = &LogEntry{time.Now()}
		mutex.Unlock()
		// pick r contacts at random, other than the sender
		// TODO: don't back-propagate, exclude all previous senders
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
}

func cleanLog(ctx context.Context) {
	ticker := time.NewTicker(cleanLogPeriod)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			mutex.Lock()
			for key, entry := range log {
				if entry.Time.Before(time.Now().Add(-logEntryLifetime)) {
					delete(log, key)
				}
			}
			mutex.Unlock()
		}
	}
}
