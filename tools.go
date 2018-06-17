// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/net/html"
)

// protect calls f and recovers any panics (that are not caused by
// runtime.Errors) and sends corresponding errors via err. When f has finished a
// single true is send via done.
func protect(err chan<- error, done chan<- bool, f func()) {
	defer func() {
		if p := recover(); p != nil {
			if _, ok := p.(runtime.Error); ok {
				panic(p)
			}

			err <- fmt.Errorf("%v", p)
		}
	}()

	f()
	done <- true
}

// hasAttributes returns true if the html.Attributes in a contain an attribute
// with key k and a value of v. If the attribute contains multiple value it is
// sufficient if is contained in the values.
func hasAttributeValue(a []html.Attribute, k string, v string) bool {
	vs := getAttributeValue(a, k)

	if len(vs) == 0 {
		return false
	}

	for _, va := range strings.Split(vs, " ") {
		if va == v {
			return true
		}
	}

	return false
}

// getAttributeValue returns the value of the attribute in a identified by key
// k. If the attribute contains multiple values all of them are returned as a
// single string.
// An empty string is returned if the aatribute identified by key k is not
// conatined in a.
func getAttributeValue(a []html.Attribute, k string) string {
	for _, at := range a {
		if at.Key == k {
			return at.Val
		}
	}

	return ""
}

// A node embeds an *html.Node and adds some convenience functions that are more
// robust against formatting changes in the HTML source.
type node struct {
	*html.Node
}

// NewNode creates and initializes a new *node embedding n. If n is nil, the new
// *node will also be nil.
func newNode(n *html.Node) *node {
	if n == nil {
		return nil
	}

	return &node{n}
}

// firstNonEmptyChild returns the first non empty child of n. "Non empty" means
// any node whose embedded *html.Node.Data consists not only of whitespace, as
// defined by Unicode.
func (n *node) firstNonEmptyChild() *node {
	if n == nil {
		return nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if len(strings.TrimSpace(c.Data)) > 0 {
			return &node{c}
		}
	}

	return nil
}

// nextNonEmptySibling returns the next non empty sibling of n. "Non empty"
// means any node whose embedded *html.Node.Data consists not only of
// whitespace, as defined by Unicode.
func (n *node) nextNonEmptySibling() *node {
	if n == nil {
		return nil
	}

	for s := n.NextSibling; s != nil; s = s.NextSibling {
		if len(strings.TrimSpace(s.Data)) > 0 {
			return &node{s}
		}
	}

	return nil
}
