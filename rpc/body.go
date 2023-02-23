package rpc

import (
	"time"
)

type (
	StoreBody struct {
		Key      string
		Value    []byte
		Modified time.Time
	}

	ListAddrBody struct {
		AddrList []string
	}
)
