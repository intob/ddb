package event

import (
	"net"

	"github.com/intob/ddb/rpc"
)

const (
	Rpc          = Topic("RPC")
	ContactAdded = Topic("CONTACT_ADDED")
)

type Topic string

type Event struct {
	Topic Topic
	Rpc   *rpc.Rpc
	Addr  *net.UDPAddr
}
