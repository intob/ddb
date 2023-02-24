package ctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/intob/ddb/event"
	"github.com/intob/ddb/id"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

type PingReq struct {
	Addr string `json:"addr"`
}

func init() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		pingReq := &PingReq{}
		err = json.Unmarshal(body, pingReq)
		if err != nil {
			fmt.Println(err)
			return
		}
		addr, err := net.ResolveUDPAddr("udp", pingReq.Addr)
		if err != nil {
			fmt.Println(err)
		}
		rpcId, err := rpc.RandId()
		if err != nil {
			fmt.Println(err)
			return
		}

		go subscribeToPingAck(rpcId)

		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   rpcId,
				Type: rpc.Ping,
			},
			Addr: addr,
		})
		if err != nil {
			fmt.Println("failed to send rpc:", err)
		}
	})
}

func subscribeToPingAck(rpcId *id.Id) {
	ev, _ := event.SubscribeOnce(func(e *event.Event) bool {
		return e.Topic == event.Rpc &&
			e.Rpc.Type == rpc.Ack &&
			bytes.Equal(*e.Rpc.Id, *rpcId)
	})
	go func() {
		timer := time.NewTimer(time.Second)
		select {
		case e := <-ev:
			fmt.Println("rcvd ping ACK", e.Rpc.Id)
		case <-timer.C:
			fmt.Println("timed out waiting for ping ACK")
		}
	}()
}
