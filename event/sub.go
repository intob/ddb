package event

import (
	"fmt"
	"sync"

	"github.com/intob/ddb/id"
)

const (
	subIdByteLen = 8
)

var (
	subs  = make(map[*id.Id]*Sub, 0)
	mutex = &sync.Mutex{}
)

type Sub struct {
	Filter func(event *Event) bool
	Rcvr   chan<- *Event
	Once   bool
}

func Subscribe(s *Sub) (*id.Id, error) {
	id, err := id.Rand(subIdByteLen)
	if err != nil {
		return nil, fmt.Errorf("failed to get random sub id: %w", err)
	}
	mutex.Lock()
	subs[id] = s
	mutex.Unlock()
	return id, nil
}

func Unsubscribe(id *id.Id) {
	mutex.Lock()
	close(subs[id].Rcvr)
	delete(subs, id)
	mutex.Unlock()
}

func Publish(event *Event) {
	for id, sub := range subs {
		if sub.Filter(event) {
			sub.Rcvr <- event
			if sub.Once {
				Unsubscribe(id)
			}
		}
	}
}
