package ctl

import (
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
	"github.com/intob/ddb/store"
	"github.com/intob/ddb/transport"
)

type GetReq struct {
	Addr string `json:"addr"`
	Key  string `json:"key"`
}

type GetResp struct {
	Value    string
	Modified time.Time
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

		rpcId := rpc.RandId()

		resultChan := make(chan *store.Entry)
		go subscribeToGetAck(rpcId, resultChan)

		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   rpcId,
				Type: rpc.Get,
				Body: []byte(getReq.Key),
			},
			Addr: addr,
		})
		if err != nil {
			fmt.Println("failed to send rpc:", err)
			return
		}

		result := <-resultChan
		if result != nil {
			resp, err := json.MarshalIndent(&GetResp{
				Value:    string(result.Value),
				Modified: result.Modified,
			}, "", "  ")
			if err != nil {
				fmt.Println("failed to marshal result")
			}
			w.Write(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
}

func subscribeToGetAck(rpcId *id.Id, result chan<- *store.Entry) {
	ev, _ := event.SubscribeOnce(event.RpcIdFilter(rpcId))
	go func() {
		timer := time.NewTimer(time.Second)
		select {
		case e := <-ev:
			detail, _ := e.Detail.(event.RpcDetail)
			fmt.Println("rcvd get ack", detail.Rpc.Id)
			body := &store.Entry{}
			if detail.Rpc.Body != nil {
				err := cbor.Unmarshal(detail.Rpc.Body, body)
				if err != nil {
					fmt.Println("failed to unmarshal rpc body:", err)
				}
				result <- body
			}
		case <-timer.C:
			fmt.Println("timed out waiting for get ack")
		}
		result <- nil
	}()
}
