package ctl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/intob/ddb/contact"
	"github.com/intob/ddb/event"
	"github.com/intob/ddb/id"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

type StoreReq struct {
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

		store.Set(storeReq.Key, []byte(storeReq.Value), time.Now())

		rpcBody, err := cbor.Marshal(&rpc.StoreBody{
			Key:      storeReq.Key,
			Value:    []byte(storeReq.Value),
			Modified: time.Now(),
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		c, err := contact.Rand(map[string]bool{})
		if err != nil {
			fmt.Println(err)
			return
		}

		rpcId := rpc.RandId()

		done := make(chan struct{})
		go subscribeToStoreAck(rpcId, done)

		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   rpcId,
				Type: rpc.Store,
				Body: rpcBody,
			},
			Addr: c.Addr,
		})
		if err != nil {
			fmt.Println("failed to send rpc:", err)
			return
		}

		<-done
	})
}

func subscribeToStoreAck(rpcId id.Id, done chan<- struct{}) {
	ev, _ := event.SubscribeOnce(event.RpcIdFilter(rpcId))
	go func() {
		timer := time.NewTimer(time.Second)
		select {
		case e := <-ev:
			detail, _ := e.Detail.(event.RpcDetail)
			fmt.Println("rcvd store ack", detail.Rpc.Id)
		case <-timer.C:
			fmt.Println("timed out waiting for store ack")
		}
		done <- struct{}{}
	}()
}
