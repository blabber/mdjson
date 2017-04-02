// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

// mdjson dumps a JSON representation of the latest MetalDays running order[1]
// to os.Stdout.
//
// [1]: http://www.metaldays.net/Line_up
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/blabber/mdjson"
)

const (
	// runningOrderURL is the string representation of the URL where the
	// latest running order can be found.
	runningOrderURL = "http://www.metaldays.net/Line_up"
)

func main() {
	err := writeRunningOrder(runningOrderURL, os.Stdout)
	if err != nil {
		log.Print(err)
	}
}

// writeRunningOrder parses the latest running order found at URL u and writes a
// JSON representation to w.
func writeRunningOrder(u string, w io.Writer) error {
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ro, err := mdjson.ParseRunningOrder(resp.Body)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(ro)
	if err != nil {
		return err
	}

	return nil
}
