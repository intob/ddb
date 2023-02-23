package rpc

type RpcType uint8

const (
	TYPE_ACK RpcType = iota
	TYPE_PING
	TYPE_STORE
	TYPE_GET
	TYPE_LIST_ADDR
)
