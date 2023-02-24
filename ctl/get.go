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

		rpcId, err := rpc.RandId()
		if err != nil {
			fmt.Println(err)
			return
		}

		resultChan := make(chan *store.Entry)
		go subscribeToGetAck(rpcId, resultChan)

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
	ev, _ := event.SubscribeOnce(func(e *event.Event) bool {
		return e.Topic == event.TOPIC_RPC &&
			e.Rpc.Type == rpc.TYPE_ACK &&
			bytes.Equal(*e.Rpc.Id, *rpcId)
	})
	go func() {
		timer := time.NewTimer(time.Second)
		select {
		case e := <-ev:
			fmt.Println("rcvd get ACK", e.Rpc.Id)
			body := &store.Entry{}
			if e.Rpc.Body != nil {
				err := cbor.Unmarshal(e.Rpc.Body, body)
				if err != nil {
					fmt.Println("failed to unmarshal rpc body:", err)
				}
				result <- body
			}
		case <-timer.C:
			fmt.Println("timed out waiting for get ACK")
		}
		result <- nil
	}()
}
