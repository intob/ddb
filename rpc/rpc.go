package rpc

import (
	"bytes"
	"fmt"
	"hash/fnv"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/id"
)

const (
	SUM_BYTE_LEN = 8
	ID_BYTE_LEN  = 8
	TYPE_PING    = 'P'
	TYPE_ACK     = 'A'
)

type RpcType byte

type Rpc struct {
	Id   *id.Id
	Type RpcType
	Body []byte
}

func (r RpcType) String() string {
	switch r {
	case TYPE_PING:
		return "PING"
	case TYPE_ACK:
		return "ACK"
	}
	return "UNKNOWN"
}

func PackRpc(r *Rpc) ([]byte, error) {
	// marshal payload
	b, err := cbor.Marshal(r)
	if err != nil {
		return nil, err
	}
	// calculate checksum
	h := fnv.New64()
	h.Write(b)
	bh := h.Sum(nil)

	// compose checksum & payload
	buf := make([]byte, 0)
	buf = append(buf, bh...)
	buf = append(buf, b...)
	return buf, nil
}

func UnpackRpc(r []byte) (*Rpc, error) {
	// verify length
	if len(r) <= SUM_BYTE_LEN {
		return nil, fmt.Errorf("msg shorter than %v bytes", SUM_BYTE_LEN)
	}

	// verify checksum
	msgSum := r[:SUM_BYTE_LEN]
	payload := r[SUM_BYTE_LEN:]
	h := fnv.New64()
	h.Write(payload)
	calcSum := h.Sum(nil)
	if !bytes.Equal(msgSum, calcSum) {
		return nil, fmt.Errorf("msg checksum is invalid")
	}

	// unmarshal payload
	rpc := &Rpc{}
	err := cbor.Unmarshal(payload, rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmashal msg: %w", err)
	}

	return rpc, nil
}
