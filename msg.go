package main

import (
	"hash/fnv"

	"github.com/fxamacker/cbor/v2"
)

type Msg struct {
	FromAddr     string
	Type         string
	InResponseTo []byte
	Body         []byte
}

func packMsg(m *Msg) ([]byte, error) {
	mb, err := cbor.Marshal(m)
	if err != nil {
		return nil, err
	}
	h := fnv.New128()
	h.Write(mb)
	mbh := h.Sum(nil)
	buf := make([]byte, 0)
	buf = append(buf, mbh...)
	buf = append(buf, mb...)
	return buf, nil
}
