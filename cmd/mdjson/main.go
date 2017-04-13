// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

// mdjson dumps a JSON representation of the latest MetalDays running order[1]
// to os.Stdout.
//
// The JSON representation follows the JSend specification[2].
//
// [1]: http://www.metaldays.net/Line_up
// [2]: https://labs.omniti.com/labs/jsend
package main

import (
	"encoding/json"
	"fmt"
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
	j, err := parseRunningOrder(runningOrderURL)
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(j)
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

// parseRunningOrder parses the latest running order found at URL u and returns
// a jsend representation.
//
// If something goes wrong the error is returned as error value and additionally
// encoded in the JSend structure.
func parseRunningOrder(u string) (jsend, error) {
	resp, err := http.Get(u)
	if err != nil {
		return newJsendError(err, http.StatusInternalServerError), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("%s returned %q", u, resp.Status)
		return newJsendError(err, http.StatusBadGateway), err
	}

	ro, err := mdjson.ParseRunningOrder(resp.Body)
	if err != nil {
		return newJsendError(err, http.StatusInternalServerError), err
	}

	return newJsend(ro), nil
}
