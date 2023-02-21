package transport

import (
	"bytes"
	"fmt"
	"hash/fnv"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/rpc"
)

const sumByteLen = 16

func packRpc(r *rpc.Rpc) ([]byte, error) {
	// marshal payload
	b, err := cbor.Marshal(r)
	if err != nil {
		return nil, err
	}

	// calculate checksum
	h := fnv.New128()
	h.Write(b)
	msg := h.Sum(nil)

	// append payload to checksum
	msg = append(msg, b...)

	fmt.Println(len(msg))
	return msg, nil
}

func unpackRpc(r []byte) (*rpc.Rpc, error) {
	// verify length
	if len(r) <= sumByteLen {
		return nil, fmt.Errorf("msg shorter than %v bytes", sumByteLen)
	}

	// verify checksum
	msgSum := r[:sumByteLen]
	payload := r[sumByteLen:]
	h := fnv.New128()
	h.Write(payload)
	calcSum := h.Sum(nil)
	if !bytes.Equal(msgSum, calcSum) {
		return nil, fmt.Errorf("msg checksum is invalid")
	}

	// unmarshal payload
	rpc := &rpc.Rpc{}
	err := cbor.Unmarshal(payload, rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmashal msg: %w", err)
	}

	return rpc, nil
}
