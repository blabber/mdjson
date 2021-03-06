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

var testAttributes = []html.Attribute{
	{Namespace: "SomeNamespace", Key: "FooKey", Val: "FooValue"},
	{Namespace: "", Key: "BarKey", Val: "BarValue1 BarValue2"},
	{Namespace: "", Key: "EmptyKey", Val: ""},
}

func TestHasAttributeValue(t *testing.T) {
	ts := []struct {
		key      string
		value    string
		expected bool
	}{
		{"FooKey", "FooBar", false},
		{"FooKey", "FooValue", true},
		{"BarKey", "BarValue1", true},
		{"BarKey", "BarValue2", true},
		{"BarKey", "BarValue3", false},
		{"EmptyKey", "", false}, // Empty value returns always false
		{"BazKey", "", false},   // Nonexisting key returns always false
	}

	for _, test := range ts {
		t.Run(fmt.Sprintf("%s_%s_%t", test.key, test.value, test.expected), func(t *testing.T) {
			is := hasAttributeValue(testAttributes, test.key, test.value)
			if is != test.expected {
				t.Errorf("hasAttributeValue returned unexpected value; expected: %t; is %t", test.expected, is)
			}
		})
	}
}

func TestGetAttributeValue(t *testing.T) {
	ts := []struct {
		key      string
		expected string
	}{
		{"FooKey", "FooValue"},
		{"BarKey", "BarValue1 BarValue2"},
		{"EmptyKey", ""},
		{"BazKey", ""},
	}

	for _, test := range ts {
		t.Run(fmt.Sprintf("%s_%s", test.key, test.expected), func(t *testing.T) {
			is := getAttributeValue(testAttributes, test.key)
			if is != test.expected {
				t.Errorf("getAttributeValue returned unexpected value; expected: %q; is %q", test.expected, is)
			}
		})
	}
}

func TestNewNodeNil(t *testing.T) {
	n := newNode(nil)

	if n != nil {
		t.Errorf("newNode(nil) returns not nil: %v", n)
	}
}

func TestNewNodeNotNil(t *testing.T) {
	hn := &html.Node{}
	n := newNode(hn)

	if n.Node != hn {
		t.Errorf("unexpected inner node; expected %v; is %v", hn, n.Node)
	}
}

func TestNilNodeFirstNonEmptyChild(t *testing.T) {
	n := newNode(nil)

	if is := n.firstNonEmptyChild(); is != nil {
		t.Error("firstNonEmptyChild() on nil node is not returning nil")
	}
}

func TestNilNodeNextNonEmptySibling(t *testing.T) {
	n := newNode(nil)

	if is := n.nextNonEmptySibling(); is != nil {
		t.Error("nextNonEmptySibling() on nil node is not returning nil")
	}
}

func TestProtectDone(t *testing.T) {
	done := make(chan bool)
	e := make(chan error)

	funcCalled := false

	go protect(e, done, func() {
		funcCalled = true
	})

	select {
	case <-done:
		break
	case err := <-e:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	if !funcCalled {
		t.Error("protect did not call wrapped function")
	}
}

func TestProtectError(t *testing.T) {
	done := make(chan bool)
	e := make(chan error)

	go protect(e, done, func() {
		panic("Testpanic")
	})

	select {
	case <-done:
		t.Fatal("unexpected success")
	case <-e:
		break
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
