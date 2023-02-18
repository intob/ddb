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
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

type StoreReq struct {
	Addr  string `json:"addr"`
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

		key, err := store.RandKey()
		if err != nil {
			fmt.Println(err)
			return
		}

		rpcBody, err := cbor.Marshal(&gossip.KV{
			Key:   key,
			Value: []byte(storeReq.Value),
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

		go subscribeToStoreAck(rpcId)

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

		w.Write([]byte(key))
	})
}

func subscribeToStoreAck(rpcId *id.Id) {
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
			fmt.Println("rcvd store ACK", e.Rpc.Id)
		case <-timer.C:
			fmt.Println("timed out waiting for store ACK")
			if !timer.Stop() {
				<-timer.C
			}
		}
		wg.Done()
	}(rcvEvents, wg)
	wg.Wait()
}
