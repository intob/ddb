package rpc

import (
	"github.com/intob/ddb/id"
)

const (
	idByteLen          = 16
	TYPE_ACK           = RpcType('A')
	TYPE_PING          = RpcType('P')
	TYPE_STORE         = RpcType('S')
	TYPE_GET           = RpcType('G')
	TYPE_LIST_CONTACTS = RpcType('C')
)

type RpcType byte

type Rpc struct {
	Id     *id.Id
	Type   RpcType
	Body   []byte
	SeenBy []string
}

func (r RpcType) String() string {
	switch r {
	case TYPE_ACK:
		return "ACK"
	case TYPE_PING:
		return "PING"
	case TYPE_STORE:
		return "STORE"
	}
	return "UNKNOWN"
}

func RandId() (*id.Id, error) {
	return id.Rand(idByteLen)
}
