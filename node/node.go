package node

import (
	"sync"
)

var (
	nodes []*Node
	mutex *sync.Mutex
)

type Node struct {
	Addr string
}

func init() {
	nodes = make([]*Node, 0)
	mutex = new(sync.Mutex)
}

func Put(node *Node) {
	mutex.Lock()
	nodes = append(nodes, node)
	mutex.Unlock()
}

func Get(addr string) *Node {
	for _, n := range nodes {
		if addr == n.Addr {
			return n
		}
	}
	return nil
}

func Rm(addr string) {
	newList := make([]*Node, len(nodes)-1)
	for _, n := range nodes {
		if addr != n.Addr {
			newList = append(newList, n)
		}
	}
	mutex.Lock()
	nodes = newList
	mutex.Unlock()
}
