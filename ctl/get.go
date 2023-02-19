package ctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/gossip"
	"github.com/intob/ddb/id"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"
)

type GetReq struct {
	Addr string `json:"addr"`
	Key  string `json:"key"`
}

func init() {
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		getReq := &GetReq{}
		err = json.Unmarshal(body, getReq)
		if err != nil {
			fmt.Println(err)
			return
		}

		addr, err := net.ResolveUDPAddr("udp", getReq.Addr)
		if err != nil {
			fmt.Println(err)
		}

		rpcId, err := rpc.RandId()
		if err != nil {
			fmt.Println(err)
			return
		}

		go subscribeToGetAck(rpcId)

		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   rpcId,
				Type: rpc.TYPE_GET,
				Body: []byte(getReq.Key),
			},
			Addr: addr,
		})
		if err != nil {
			fmt.Println("failed to send rpc:", err)
			return
		}
	})
}

func subscribeToGetAck(rpcId *id.Id) {
	rcvEvents := make(chan *event.Event)
	event.Subscribe(&event.Sub{
		Filter: func(e *event.Event) bool {
			return e.Topic == event.TOPIC_RPC &&
				e.Rpc.Type == rpc.TYPE_ACK &&
				bytes.Equal(*e.Rpc.Id, *rpcId)
		},
		Rcvr: rcvEvents,
		Once: true,
	})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(rcvEvents <-chan *event.Event, wg *sync.WaitGroup) {
		timer := time.NewTimer(time.Second)
		select {
		case e := <-rcvEvents:
			fmt.Println("rcvd get ACK", e.Rpc.Id)
			body := &gossip.StoreRpcBody{}
			if e.Rpc.Body != nil {
				err := cbor.Unmarshal(e.Rpc.Body, body)
				if err != nil {
					fmt.Println("failed to unmarshal rpc body:", err)
				}
				fmt.Println("value:", string(body.Value))
				fmt.Println("modified:", body.Modified)
			} else {
				fmt.Println("no value")
			}
		case <-timer.C:
			fmt.Println("timed out waiting for get ACK")
			if !timer.Stop() {
				<-timer.C
			}
		}
		wg.Done()
	}(rcvEvents, wg)
	wg.Wait()
}
