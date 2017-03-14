// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

//mdjson opens a HTTP server on port 8080 and serves a JSON representation of
//the current MetalDays running order.
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/blabber/mdjson"
)

const (
	URL = "http://www.metaldays.net/Line_up"
)

func main() {
	http.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(URL)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		ro, err := mdjson.ParseRunningOrder(resp.Body)
		if err != nil {
			panic(err)
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(ro)
		if err != nil {
			panic(err)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
