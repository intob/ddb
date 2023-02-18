package ctl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

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
