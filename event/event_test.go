package event

import (
	"testing"
	"time"
)

func TestSubscribe(t *testing.T) {
	UnsubscribeAll()
	ev, subId := Subscribe(func(e *Event) bool {
		return true
	})
	done := make(chan struct{})
	go func() {
		i := 0
		for range ev {
			i++
		}
		if i == 3 {
			done <- struct{}{}
		}
	}()
	Publish(&Event{})
	Publish(&Event{})
	Publish(&Event{})
	Unsubscribe(subId)
	timer := time.NewTimer(time.Millisecond)
	select {
	case <-timer.C:
		t.Fatalf("timeout")
	case <-done:
		break
	}
}
