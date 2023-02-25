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
	logEntryLifetime = 10 * time.Second
	cleanLogPeriod   = 10 * time.Second
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
		return e.Topic == event.Rpc &&
			e.Rpc.Type == rpc.Store
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
		contacts := make([]*contact.Contact, 0)
		exclude := map[string]bool{e.Addr.String(): true}
		for i := 0; i < r; i++ {
			rnd, err := contact.Rand(exclude)
			if err != nil {
				fmt.Println("failed to pick random contact:", err)
				break
			}
			contacts = append(contacts, rnd)
			exclude[rnd.Addr.String()] = true
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
