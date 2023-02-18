package event

import "github.com/intob/ddb/rpc"

const (
	TOPIC_RPC         = "RPC"
	EVENT_ID_BYTE_LEN = 8
)

type Event struct {
	Topic string
	Id    [EVENT_ID_BYTE_LEN]byte
	Rpc   *rpc.Rpc
}
