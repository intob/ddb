package ctl

import (
	"fmt"
	"net/http"
	"os"
)

var ctlAddr = ""

func init() {
	// parse args
	for i, arg := range os.Args {
		if arg == "--ctl" && len(os.Args) > i+1 {
			ctlAddr = os.Args[i+1]
		}
	}
}

func Start() error {
	fmt.Printf("ctl server listening on %s\r\n", ctlAddr)
	err := http.ListenAndServe(ctlAddr, nil)
	if err != nil {
		return err
	}
	return nil
}
