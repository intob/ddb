package ctl

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/intob/ddb/contact"
)

type Info struct {
	Contacts int `json:"contacts"`
}

func init() {
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		i, err := json.MarshalIndent(&Info{
			Contacts: contact.Count(),
		}, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		w.Write(i)
	})
}
