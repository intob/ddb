package event

import "github.com/intob/ddb/rpc"

const (
	TOPIC_RPC         = "RPC"
	EVENT_ID_BYTE_LEN = 8
)

type Event struct {
	Id    []byte
	Topic string
	Rpc   *rpc.Rpc
}
