package rpc

type RpcType uint8

const (
	Ack RpcType = iota
	Ping
	Store
	Get
	ListAddr
)
