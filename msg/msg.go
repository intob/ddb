package msg

import (
	"bytes"
	"fmt"
	"hash/fnv"

	"github.com/fxamacker/cbor/v2"
)

const MSG_SUM_BYTE_LENGTH = 8

type Msg struct {
	ReplyAddr    string
	Type         string
	InResponseTo []byte
	Body         []byte
}

func PackMsg(m *Msg) ([]byte, error) {
	b, err := cbor.Marshal(m)
	if err != nil {
		return nil, err
	}
	h := fnv.New64()
	h.Write(b)
	bh := h.Sum(nil)
	buf := make([]byte, 0)
	buf = append(buf, bh...)
	buf = append(buf, b...)
	return buf, nil
}

func UnpackMsg(m []byte) (*Msg, error) {
	if len(m) <= MSG_SUM_BYTE_LENGTH {
		return nil, fmt.Errorf("msg shorter than %v bytes", MSG_SUM_BYTE_LENGTH)
	}

	// verify checksum
	msgSum := m[:MSG_SUM_BYTE_LENGTH]
	payload := m[MSG_SUM_BYTE_LENGTH:]
	h := fnv.New64()
	h.Write(payload)
	calcSum := h.Sum(nil)
	if !bytes.Equal(msgSum, calcSum) {
		return nil, fmt.Errorf("msg checksum is invalid")
	}

	// decode payload
	msg := &Msg{}
	err := cbor.Unmarshal(payload, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmashal msg: %w", err)
	}

	return msg, nil
}
