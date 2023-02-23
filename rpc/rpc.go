package rpc

import (
	"github.com/intob/ddb/id"
)

const idByteLen = 16

type Rpc struct {
	Id     *id.Id
	Type   RpcType
	Body   []byte
	SeenBy []string
}

func RandId() (*id.Id, error) {
	return id.Rand(idByteLen)
}
