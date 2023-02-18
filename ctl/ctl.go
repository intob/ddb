package ctl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/intob/ddb/event"
	"github.com/intob/ddb/id"
	"github.com/intob/ddb/rpc"
	"github.com/intob/ddb/transport"

	"github.com/intob/ddb/contact"
)

var ctlAddr = ""

type Info struct {
	Contacts int `json:"contacts"`
}

type PingReq struct {
	Addr string `json:"addr"`
}

func init() {
	// parse args
	for i, arg := range os.Args {
		if arg == "--ctl" && len(os.Args) > i+1 {
			ctlAddr = os.Args[i+1]
		}
	}

	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		i, err := json.MarshalIndent(&Info{
			Contacts: contact.Count(),
		}, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		w.Write(i)
	})

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

		go subscribeToAck(rpcId)

		err = transport.SendRpc(&transport.AddrRpc{
			Rpc: &rpc.Rpc{
				Id:   rpcId,
				Type: rpc.TYPE_PING,
			},
			Addr: addr,
		})
		if err != nil {
			fmt.Println("failed to send rpc:", err)
		}
	})
}

func Start(ctx context.Context) error {
	fmt.Printf("ctl server listening on %s\r\n", ctlAddr)
	err := http.ListenAndServe(ctlAddr, nil)
	if err != nil {
		return err
	}
	return nil
}

func subscribeToAck(rpcId *id.Id) {
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
			fmt.Println("rcvd ping ACK", e.Rpc.Id)
		case <-timer.C:
			fmt.Println("timed out waiting for ACK")
			if !timer.Stop() {
				<-timer.C
			}
		}
		wg.Done()
	}(rcvEvents, wg)
	wg.Wait()
	fmt.Println("sub done")
}
