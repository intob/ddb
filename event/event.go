package event

import (
	"bytes"
	"net"

	"github.com/intob/ddb/id"
	"github.com/intob/ddb/rpc"
)

const (
	Rpc          = Topic("RPC")
	ContactAdded = Topic("CONTACT_ADDED")
)

type Topic string

type Event struct {
	Topic  Topic
	Detail any
}

type ContactAddedDetail struct {
	Addr *net.UDPAddr
}

type RpcDetail struct {
	Rpc  *rpc.Rpc
	Addr *net.UDPAddr
}

func RpcTypeFilter(typ rpc.RpcType) func(e *Event) bool {
	return func(e *Event) bool {
		if e.Topic != Rpc {
			return false
		}
		detail, ok := e.Detail.(RpcDetail)
		if !ok {
			return false
		}
		return detail.Rpc.Type == typ
	}
}

func RpcIdFilter(id *id.Id) func(e *Event) bool {
	return func(e *Event) bool {
		if e.Topic != Rpc {
			return false
		}
		detail, ok := e.Detail.(RpcDetail)
		if !ok {
			return false
		}
		return bytes.Equal(*detail.Rpc.Id, *id)
	}
}
