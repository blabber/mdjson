// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	testdataValidHTML = "../../testdata/sample.html"
	testdataValidJSON = "../../testdata/sample.json"

	testdataInvalidHTML = "../../testdata/fail.html"
	testdataInvalidJSON = "../../testdata/fail.json"

	messagePrefixParseError = "Unable to parse running order structure "
)

func messageSuffixRemoteError(c int) string {
	return fmt.Sprintf(" returned \"%d %s\"", c, http.StatusText(c))
}

func init() {
	log.SetOutput(ioutil.Discard)
}

func dataHandler(dr io.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := io.Copy(w, dr)
		if err != nil {
			panic(err)
		}
	}
}

func codeHandler(c int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(c)
		_, err := w.Write([]byte(http.StatusText(c)))
		if err != nil {
			panic(err)
		}
	}
}

var dataTests = []struct {
	name         string
	inputData    string
	expectedData string
	validData    bool
	cors         bool
}{
	{"valid_without_cors", testdataValidHTML, testdataValidJSON, true, false},
	{"invalid_without_cors", testdataInvalidHTML, testdataInvalidJSON, false, false},
	{"valid_with_cors", testdataValidHTML, testdataValidJSON, true, true},
	{"invalid_with_cors", testdataInvalidHTML, testdataInvalidJSON, false, true},
}

func TestDumpData(t *testing.T) {
	for _, dt := range dataTests {
		t.Run(dt.name, func(t *testing.T) {
			f, err := os.Open(dt.inputData)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			s := httptest.NewServer(dataHandler(f))
			defer s.Close()

			var b bytes.Buffer
			err = dump(s.URL, &b)
			if err != nil {
				if dt.validData {
					t.Fatal(err)
				} else {
					expectedPrefix := messagePrefixParseError
					is := err.Error()
					if !strings.HasPrefix(is, expectedPrefix) {
						t.Fatalf("unexpected error; expected: \"...%s\"; is: %q", expectedPrefix, is)
					}
				}
			} else if !dt.validData {
				t.Error("expected error did not occur")
			}
			is := b.Bytes()

			expected, err := ioutil.ReadFile(dt.expectedData)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(is, expected) {
				t.Errorf("dump wrote unexpected data; expected: %q; is: %q", expected, is)
			}
		})
	}
}

func TestServeData(t *testing.T) {
	for _, dt := range dataTests {
		t.Run(dt.name, func(t *testing.T) {
			f, err := os.Open(dt.inputData)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			s := httptest.NewServer(dataHandler(f))
			defer s.Close()

			rw := httptest.NewRecorder()
			rr, err := http.NewRequest("GET", "http://example.com/runningorder.json", nil)
			if err != nil {
				t.Fatal(err)
			}

			h := runningorderHandler(s.URL, flags{cors: &dt.cors})
			h(rw, rr)

			isACAOHeader := rw.HeaderMap.Get("Access-Control-Allow-Origin")
			expectedACAOHeader := ""
			if dt.cors {
				expectedACAOHeader = "*"
			}
			if isACAOHeader != expectedACAOHeader {
				t.Errorf("unexpected Access-Control-Allow-Origin header; expected %q; is: %q", expectedACAOHeader, isACAOHeader)
			}

			r := rw.Result()

			isCode := r.StatusCode
			expectedCode := http.StatusOK
			if !dt.validData {
				expectedCode = http.StatusInternalServerError
			}
			if isCode != expectedCode {
				t.Errorf("unexpected status; expected: %d; is: %d", expectedCode, isCode)
			}

			is, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}

			expected, err := ioutil.ReadFile(dt.expectedData)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal([]byte(is), expected) {
				t.Errorf("response contained unexpected data; expected: %q; is: %q", expected, is)
			}
		})
	}
}

var remoteErrorTests = []struct {
	name string
	code int
	cors bool
}{
	{"404_with_cors", 404, true},
	{"404_without_cors", 404, false},
	{"500_with_cors", 500, true},
	{"500_without_cors", 500, false},
}

func TestDumpRemoteError(t *testing.T) {
	for _, ret := range remoteErrorTests {
		t.Run(ret.name, func(t *testing.T) {
			s := httptest.NewServer(codeHandler(ret.code))
			defer s.Close()

			var b bytes.Buffer
			err := dump(s.URL, &b)
			if err != nil {
				expectedSuffix := messageSuffixRemoteError(ret.code)
				is := err.Error()
				if !strings.HasSuffix(is, expectedSuffix) {
					t.Errorf("unexpected error; expected: '...%s'; is: %s", expectedSuffix, is)
				}
			} else {
				t.Error("expected error did not occur")
			}

			var js jsend
			dec := json.NewDecoder(&b)
			err = dec.Decode(&js)
			if err != nil {
				t.Fatal(err)
			}

			checkFailedJsend(ret.code, js, t)
		})
	}
}

func TestServeRemoteError(t *testing.T) {
	for _, ret := range remoteErrorTests {
		t.Run(ret.name, func(t *testing.T) {
			s := httptest.NewServer(codeHandler(ret.code))
			defer s.Close()

			rw := httptest.NewRecorder()
			rr, err := http.NewRequest("GET", "http://example.com/runningorder.json", nil)
			if err != nil {
				t.Fatal(err)
			}

			h := runningorderHandler(s.URL, flags{cors: &ret.cors})
			h(rw, rr)

			isACAOHeader := rw.HeaderMap.Get("Access-Control-Allow-Origin")
			expectedACAOHeader := ""
			if ret.cors {
				expectedACAOHeader = "*"
			}
			if isACAOHeader != expectedACAOHeader {
				t.Errorf("unexpected Access-Control-Allow-Origin header; expected %q; is: %q", expectedACAOHeader, isACAOHeader)
			}

			r := rw.Result()

			expectedCode := http.StatusBadGateway
			isCode := r.StatusCode
			if isCode != expectedCode {
				t.Errorf("unexpected status; expected: %d; is: %d", expectedCode, isCode)
			}

			var js jsend
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&js)
			if err != nil {
				t.Fatal(err)
			}

			checkFailedJsend(ret.code, js, t)
		})
	}
}

func checkFailedJsend(c int, j jsend, t *testing.T) {
	isStatus := j.Status
	expectedStatus := "error"
	if isStatus != expectedStatus {
		t.Errorf("unexpected jsend.Status; expected: %q; is: %q", expectedStatus, isStatus)
	}

	expectedCode := http.StatusBadGateway
	isCode := j.Code
	if isCode != expectedCode {
		t.Errorf("unexpected jsend.Code; expected: %d; is: %d", expectedCode, isCode)
	}

	expectedSuffix := messageSuffixRemoteError(c)
	is := j.Message
	if !strings.HasSuffix(is, expectedSuffix) {
		t.Errorf("unexpected jsend.Message; expected: '...%s'; is: %s", expectedSuffix, is)
	}
}
