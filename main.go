package main

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"net"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/node"
	"github.com/intob/ddb/store"
)

var laddr *net.UDPAddr

func main() {
	var err error
	laddr, err = net.ResolveUDPAddr("udp", ":1992")
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}
		fmt.Println("Received: ", string(buf[:n]), " from ", raddr)

		resp, err := handleIncoming(buf[:n])
		if err != nil {
			fmt.Println("Error: ", err)
		}
		if resp != nil {
			packed, err := packMsg(resp)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			conn.WriteToUDP(packed, raddr)
		}
	}
}

func handleIncoming(msg []byte) (*Msg, error) {
	if len(msg) <= 16 {
		return &Msg{
			Type: "ERROR",
			Body: []byte("invalid message"),
		}, nil
	}

	// verify checksum
	msgsum := msg[:16]
	payload := msg[16:]
	h := fnv.New128()
	h.Write(payload)
	sum := h.Sum(nil)
	if !bytes.Equal(msgsum, sum) {
		return &Msg{
			Type:         "ERROR",
			InResponseTo: msgsum,
			Body:         []byte("payload did not match checksum"),
		}, nil
	}

	// decode payload
	m := &Msg{}
	err := cbor.Unmarshal(payload, m)
	if err != nil {
		return &Msg{
			Type:         "ERROR",
			InResponseTo: msgsum,
			Body:         []byte("failed to unmarshal payload"),
		}, err
	}

	if node.Get(m.FromAddr) == nil {
		node.Put(&node.Node{
			Addr: m.FromAddr,
		})
		fmt.Println("added to node list:", m.FromAddr)
	}

	resp, err := handleDecodedMsg(m)
	if resp != nil {
		resp.FromAddr = laddr.String()
		resp.InResponseTo = msgsum
	}
	return resp, err
}

func handleDecodedMsg(m *Msg) (*Msg, error) {
	switch m.Type {
	case "PING":
		return &Msg{
			Type: "ACK",
		}, nil

	case "SET":
		kv := &store.KV{}
		err := cbor.Unmarshal(m.Body, kv)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal SET msg: %w", err)
		}
		fmt.Println("set: ", kv)
		store.Set(kv.Key, kv.Value)
		return &Msg{
			Type: "ACK",
			Body: []byte("set"),
		}, nil

	case "GET":
		v := store.Get(string(m.Body))
		return &Msg{
			Type: "VALUE",
			Body: v,
		}, nil
	}
	return &Msg{
		Type: "ERROR",
		Body: []byte("unknown message type"),
	}, nil
}
