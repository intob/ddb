package event

import (
	"github.com/intob/ddb/rpc"
)

const (
	TOPIC_RPC = "RPC"
)

type Event struct {
	Topic string
	Rpc   *rpc.Rpc
}
