package event

import (
	"net"

	"github.com/intob/ddb/rpc"
)

const (
	TOPIC_RPC           = Topic("RPC")
	TOPIC_CONTACT_ADDED = Topic("CONTACT_ADDED")
)

type Topic string

type Event struct {
	Topic Topic
	Rpc   *rpc.Rpc
	Addr  *net.UDPAddr
}
