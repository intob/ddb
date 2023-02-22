package transport

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"hash/fnv"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/rpc"
)

const sumByteLen = 16

func packRpc(r *rpc.Rpc) ([]byte, error) {
	// marshal & gzip payload
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	defer gz.Close()

	cb := cbor.NewEncoder(gz)
	err := cb.Encode(r)
	if err != nil {
		return nil, err
	}

	gz.Flush()
	bufBytes := buf.Bytes()

	// calculate checksum
	h := fnv.New128()
	_, err = h.Write(bufBytes)
	if err != nil {
		return nil, err
	}
	msg := h.Sum(nil)

	// append payload to checksum
	msg = append(msg, bufBytes...)

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
	_, err := h.Write(payload)
	if err != nil {
		return nil, err
	}
	calcSum := h.Sum(nil)
	if !bytes.Equal(msgSum, calcSum) {
		return nil, fmt.Errorf("msg checksum is invalid")
	}

	// ungzip & unmarshal payload
	buf := bytes.NewReader(payload)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	cb := cbor.NewDecoder(gz)
	rpc := &rpc.Rpc{}
	err = cb.Decode(rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cbor: %w", err)
	}

	return rpc, nil
}
