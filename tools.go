// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"strings"

	"golang.org/x/net/html"
)

// hasAttributes returns true if the html.Attributes in a contain an attribute
// with key k and a value of v.  If the attribute contains multiple value it is
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
