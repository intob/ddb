package ctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/id"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/subs"
	"github.com/intob/ddb/transport"
)

type StoreReq struct {
	Addr  string `json:"addr"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func init() {
	http.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		storeReq := &StoreReq{}
		err = json.Unmarshal(body, storeReq)
		if err != nil {
			fmt.Println(err)
			return
		}

		rpcBody, err := cbor.Marshal(&subs.StoreRpcBody{
			Key:      storeReq.Key,
			Value:    []byte(storeReq.Value),
			Modified: time.Now(),
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		addr, err := net.ResolveUDPAddr("udp", storeReq.Addr)
		if err != nil {
			fmt.Println(err)
		}

		rpcId, err := rpc.RandId()
		if err != nil {
			fmt.Println(err)
			return
		}

		done := make(chan struct{})
		go subscribeToStoreAck(rpcId, done)

		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   rpcId,
				Type: rpc.TYPE_STORE,
				Body: rpcBody,
			},
			Addr: addr,
		})
		if err != nil {
			fmt.Println("failed to send rpc:", err)
			return
		}

		<-done
		fmt.Println("store done")
	})
}

func subscribeToStoreAck(rpcId *id.Id, done chan<- struct{}) {
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
	go func() {
		timer := time.NewTimer(time.Second)
		select {
		case e := <-rcvEvents:
			fmt.Println("rcvd store ACK", e.Rpc.Id)
		case <-timer.C:
			fmt.Println("timed out waiting for store ACK")
		}
		fmt.Println("will close done chan")
		done <- struct{}{}
	}()
}
