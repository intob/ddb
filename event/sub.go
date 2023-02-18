package event

import "fmt"

const (
	SUB_ID_BYTE_LEN = 8
)

var (
	subs = make([]*Sub, 0)
)

type Sub struct {
	Id        []byte
	MatchFunc func(event *Event) bool
	Rcvr      chan<- *Event
}

func init() {
	fmt.Println("array state", subs)
}

func Subscribe(s *Sub) {
	subs = append(subs, s)
}

func Publish(event *Event) {
	for _, sub := range subs {
		if sub.MatchFunc(event) {
			sub.Rcvr <- event
		}
	}
}
