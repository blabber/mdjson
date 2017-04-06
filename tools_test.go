// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"testing"

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
		is := hasAttributeValue(testAttributes, test.key, test.value)
		if is != test.expected {
			t.Errorf("hasAttributeValue(testAttributes, \"%s\", \"%s\") returns %t; expected %t",
				test.key, test.value, is, test.expected)
		}
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
		is := getAttributeValue(testAttributes, test.key)
		if is != test.expected {
			t.Errorf("getAttributeValue(testAttributes, \"%s\") returns \"%s\"; expected \"%s\"",
				test.key, is, test.expected)
		}
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
