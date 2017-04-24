// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func sampleDataHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("../../testdata/sample.html")
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(w, f)
	if err != nil {
		panic(err)
	}
}

func TestDumpValidData(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(sampleDataHandler))
	defer s.Close()

	var b bytes.Buffer
	err := dump(s.URL, &b)
	if err != nil {
		t.Fatal(err)
	}
	is := b.Bytes()

	expected, err := ioutil.ReadFile("../../testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(is, expected) != 0 {
		t.Errorf("dump wrote unexpected data; expected: %q; is: %q", expected, is)
	}
}
