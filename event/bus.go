package event

import "fmt"

var (
	subs = make([]*Sub, 0)
)

type Sub struct {
}

func init() {
	fmt.Println("array state", subs)
}

func Subscribe(s *Sub) {
	subs = append(subs, s)
}
