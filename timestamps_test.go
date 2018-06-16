// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func compareTimeStampsPointers(is, expected *TimeStamps) func(*testing.T) {
	return func(t *testing.T) {
		if is != nil && expected != nil && *is != *expected {
			t.Errorf("Unexpected timestamps; is \"%v\"; expected \"%v\"", *is, *expected)
		}

		if (is == nil && expected != nil) || (is != nil && expected == nil) {
			t.Errorf("Unexpected timestamps; is \"%p\"; expected \"%p\"", is, expected)
		}
	}
}
