package event

import (
	"sync"

	"github.com/intob/ddb/id"
)

const (
	subIdByteLen = 8
)

var (
	subs  = make(map[*id.Id]*sub, 0)
	mutex = &sync.Mutex{}
)

type sub struct {
	filter func(e *Event) bool
	rcvr   chan *Event
	once   bool
}

func Subscribe(filter func(e *Event) bool) (<-chan *Event, *id.Id) {
	return subscribe(filter, false)
}

func SubscribeOnce(filter func(e *Event) bool) (<-chan *Event, *id.Id) {
	return subscribe(filter, true)
}

func subscribe(filter func(e *Event) bool, once bool) (<-chan *Event, *id.Id) {
	sub := &sub{
		filter: filter,
		rcvr:   make(chan *Event),
		once:   once,
	}
	id := id.Rand(subIdByteLen)
	mutex.Lock()
	subs[id] = sub
	mutex.Unlock()
	return sub.rcvr, id
}

func Unsubscribe(id *id.Id) {
	mutex.Lock()
	close(subs[id].rcvr)
	delete(subs, id)
	mutex.Unlock()
}

func Publish(event *Event) {
	mutex.Lock()
	for id, sub := range subs {
		if sub.filter(event) {
			sub.rcvr <- event
			if sub.once {
				close(sub.rcvr)
				delete(subs, id)
			}
		}
	}
	mutex.Unlock()
}

func UnsubscribeAll() {
	mutex.Lock()
	for id, sub := range subs {
		close(sub.rcvr)
		delete(subs, id)
	}
	mutex.Unlock()
}
