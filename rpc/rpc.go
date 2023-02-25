package rpc

import (
	"github.com/intob/ddb/id"
)

const idByteLen = 16

type Rpc struct {
	Id   *id.Id
	Type RpcType
	Body []byte
}

func RandId() *id.Id {
	return id.Rand(idByteLen)
}
