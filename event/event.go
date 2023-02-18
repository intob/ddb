package event

import (
	"net"

	"github.com/intob/ddb/rpc"
)

const (
	TOPIC_RPC = "RPC"
)

type Event struct {
	Topic string
	Rpc   *rpc.Rpc
	Raddr *net.UDPAddr
}
