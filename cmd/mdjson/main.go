// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

// mdjson scrapes the latest MetalDays running order[1] and provides a JSON
// representation of the running order. The JSON representation follows the
// JSend specification[2].
//
// By default mdjson just dumps the running order to os.Stdout, but you can turn
// it into a HTTP server by providing the -http flag. If you start mdjson as
// follows
//
//	mdjson -http=":8080"
//
// you can access the running order on port 8080. The path under which the JSON
// is served is "/runningorder.json". Using curl you can access the running
// order by calling
//
//	curl "http://localhost:8080/runningorder.json"
//
// You can tell the HTTP server to add a wildcard Access-Control-Allow-Origin
// header to the replies by providing the -cors flag.
//
// [1]: http://www.metaldays.net/Line_up
// [2]: https://labs.omniti.com/labs/jsend
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/blabber/mdjson"
)

type flags struct {
	http *string
	cors *bool
}

func main() {
	// runningOrderURL is the string representation of the URL where the
	// latest running order can be found.
	const runningOrderURL = "http://www.metaldays.net/Line_up"

	var flags = flags{
		http: flag.String("http", "", "HTTP service address"),
		cors: flag.Bool("cors", false, "add wildcard Access-Control-Allow-Origin header to HTTP replies"),
	}

	flag.Parse()

	if len(*flags.http) > 0 {
		log.Fatal(serve(runningOrderURL, flags))
	}

	err := dump(runningOrderURL, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

// jsend is a simple envelope for a JSend compliant structure. The JSend
// specification can be found here: https://labs.omniti.com/labs/jsend
type jsend struct {
	// Status contains the return status of an API call. The JSend
	// specification defines the following values:
	//
	// 	"success": The API call was successful. The returned data is
	// 		contained in the Data field.
	//	"fail":	The API call failed due to some invalid data or call
	//	 	conditions. The field Data is expected to contain an
	//	 	object that explains why the call failed. This tool
	//	 	never returns the "fail" status as we are not processing
	//	 	any input data and the API we use is stateless.
	//	"error": The API call failed because of an issue in the backend.
	//		The field Message contains a description of the issue.
	//		The field code may optionally contain a numeric error
	//		code. The JSend specification defines that the field
	//		Data may optionally contain additional data about the
	//		error that occured, but mdjson is not using this field
	//		for error data.
	Status string `json:"status,omitempty"`

	// Data contains the parsed running order if Status is "success".
	Data *mdjson.RunningOrder `json:"data,omitempty"`

	// Message contains a human readable error message if Status is "error".
	Message string `json:"message,omitempty"`

	// Code may contain a numeric error code if Status is "error". mdjson
	// uses this field for the HTTP status code describing the error that
	// occured.
	Code int `json:"code,omitempty"`
}

// newJsend initializes a new jsend with Status "success" and ro as Data.
func newJsend(ro *mdjson.RunningOrder) jsend {
	return jsend{
		Status: "success",
		Data:   ro,
	}
}

// newJsendError initializes a new jsend with Status "error" and a string
// representation of err as Message and code as Code.
func newJsendError(err error, code int) jsend {
	return jsend{
		Status:  "error",
		Message: fmt.Sprintf("%v", err),
		Code:    code,
	}
}

// serve starts a HTTP server listening at address flags.http. It serves a JSON
// representation of the latest running order, found at URL u, under path
// "/runningorder.json".
func serve(u string, flags flags) error {
	http.Handle("/runningorder.json", runningorderHandler(u, flags))

	return http.ListenAndServe(*flags.http, nil)
}

// runningorderHandler returns a http.HandlerFunc that serves a JSON
// representation of the latest running order found at URL u.
//
// If flags.cors is true, a wildcard Access-Control-Allow-Origin is added to the
// response.
func runningorderHandler(u string, flags flags) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("running order request received")

		w.Header().Set("Content-Type", "application/json")
		if *flags.cors {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		j, err := parseRunningOrder(u)
		if err != nil {
			log.Printf("parseRunningorder: %v", err)
			w.WriteHeader(j.Code)
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(j)
		if err != nil {
			log.Printf("encode: %v", err)
		}
	}
}

// dump parses the latest running order found at URL u and writes a JSON
// representation to w.
func dump(u string, w io.Writer) error {
	j, parseErr := parseRunningOrder(u)

	enc := json.NewEncoder(w)
	err := enc.Encode(j)
	if err != nil {
		return err
	}

	if parseErr != nil {
		return parseErr
	}

	return nil
}

// parseRunningOrder parses the latest running order found at URL u and returns
// a jsend representation.
//
// If something goes wrong the error is returned as error value and additionally
// encoded in the JSend structure.
func parseRunningOrder(u string) (jsend, error) {
	resp, err := http.Get(u)
	if err != nil {
		return newJsendError(err, http.StatusBadGateway), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s returned %q", u, resp.Status)
		return newJsendError(err, http.StatusBadGateway), err
	}

	ro, err := mdjson.ParseRunningOrder(resp.Body)
	if err != nil {
		return newJsendError(err, http.StatusInternalServerError), err
	}

	return newJsend(ro), nil
}
