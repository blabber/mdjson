// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package main

import (
	"bytes"
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
)

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
		w.Write([]byte(http.StatusText(c)))
	}
}

func TestDumpValidData(t *testing.T) {
	f, err := os.Open(testdataValidHTML)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := httptest.NewServer(dataHandler(f))
	defer s.Close()

	var b bytes.Buffer
	err = dump(s.URL, &b)
	if err != nil {
		t.Fatal(err)
	}
	is := b.Bytes()

	expected, err := ioutil.ReadFile(testdataValidJSON)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(is, expected) != 0 {
		t.Errorf("dump wrote unexpected data; expected: %q; is: %q", expected, is)
	}
}

func TestDumpInvalidData(t *testing.T) {
	f, err := os.Open(testdataInvalidHTML)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := httptest.NewServer(dataHandler(f))
	defer s.Close()

	var b bytes.Buffer
	err = dump(s.URL, &b)
	if err == nil {
		t.Error("expected error did not occur")
	}
	if err != nil && !strings.HasPrefix(err.Error(), "Unable to parse running order structure ") {
		t.Errorf("Unexpected error; expected: \"Unable to parse running order structure...\"; is: %v", err)
	}
}

func TestDumpRemoteError(t *testing.T) {
	c := 404
	s := httptest.NewServer(codeHandler(c))
	defer s.Close()

	var b bytes.Buffer
	err := dump(s.URL, &b)
	if err == nil {
		t.Error("expected error did not occur")
	}

	st := http.StatusText(c)
	expectedSuffix := fmt.Sprintf(" returned \"%d %s\"", c, st)
	if err != nil && !strings.HasSuffix(err.Error(), expectedSuffix) {
		t.Errorf("Unexpected error; expected: '...%s'; is: %v", expectedSuffix, err)
	}
}

func TestServeValidData(t *testing.T) {
	f, err := os.Open(testdataValidHTML)
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

	c := false
	h := runningorderHandler(s.URL, flags{cors: &c})
	h(rw, rr)

	if hv := rw.HeaderMap.Get("Access-Control-Allow-Origin"); hv != "" {
		t.Errorf("unexpected Access-Control-Allow-Origin header: %q", hv)
	}

	r := rw.Result()

	expectedCode := http.StatusOK
	isCode := r.StatusCode
	if isCode != expectedCode {
		t.Errorf("unexpected status; expected: %d; is: %d", expectedCode, isCode)
	}

	is, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := ioutil.ReadFile(testdataValidJSON)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare([]byte(is), expected) != 0 {
		t.Errorf("response contained unexpected data; expected: %q; is: %q", expected, is)
	}
}

func TestServeInvalidData(t *testing.T) {
	f, err := os.Open(testdataInvalidHTML)
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

	c := false
	h := runningorderHandler(s.URL, flags{cors: &c})
	h(rw, rr)

	if hv := rw.HeaderMap.Get("Access-Control-Allow-Origin"); hv != "" {
		t.Errorf("unexpected Access-Control-Allow-Origin header: %q", hv)
	}

	r := rw.Result()

	expectedCode := http.StatusInternalServerError
	isCode := r.StatusCode
	if isCode != expectedCode {
		t.Errorf("unexpected status; expected: %d; is: %d", expectedCode, isCode)
	}

	is, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := ioutil.ReadFile(testdataInvalidJSON)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare([]byte(is), expected) != 0 {
		t.Errorf("response contained unexpected data; expected: %q; is: %q", expected, is)
	}
}
