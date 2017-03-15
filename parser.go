// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

// A Day contains a label (ideally the date of the day), the stages that have
// associated events for the day and the *html.Node representing the day.
type Day struct {
	Label  string
	Stages []Stage
	node   *html.Node
}

// getDays walks the running order starting at n and sends any Day found via d.
// d is closed once GetDays has finished its job.
func getDays(n *html.Node, d chan<- Day) {
	getDaysRecursive(n, d)
	close(d)
}

// getDaysRecursive is used by GetDays (and by itself) to walk the running
// order recursively starting at n. Any Day found is published via d.
func getDaysRecursive(n *html.Node, d chan<- Day) {
	if n.Type == html.ElementNode && hasAttributeValue(n.Attr, "class", "lineup_day") {
		nn := node{n}
		date := nn.firstNonEmptyChild().nextNonEmptySibling().firstNonEmptyChild().FirstChild.Data
		// For some reason there is an additional space behind each date separator
		date = strings.Replace(date, ". ", ".", -1)
		d <- Day{date, []Stage{}, n}
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getDaysRecursive(c, d)
	}
}

// A Stage contains a label (the name of the stage), the events associated with
// the stage and the *html.Node representing the stage.
type Stage struct {
	Label  string
	Events []Event
	node   *html.Node
}

// getStages walks the running order starting at n and sends any Stage found
// via s. s is closed once GetStages has finished its job. In order to get the
// stages for one day, n should be the node associated with a Day.
func getStages(n *html.Node, s chan<- Stage) {
	getStagesRecursive(n, s)
	close(s)
}

// getStagesRecursive is used by GetStages (and by itself) to walk the running
// order recursively starting at n. Any Stage found is published via s.
func getStagesRecursive(n *html.Node, s chan<- Stage) {
	if n.Type == html.ElementNode && hasAttributeValue(n.Attr, "class", "lineup_stage") {
		nn := node{n}
		name := nn.firstNonEmptyChild().firstNonEmptyChild().nextNonEmptySibling().FirstChild.Data
		s <- Stage{strings.Title(name), []Event{}, n}
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getStagesRecursive(c, s)
	}
}

// A event contains a Label (the band playing) and the node representing the
// event.
type Event struct {
	Label string
	node  *html.Node
}

// getEvents walks the running order starting at n and sends any Event found
// via e. e is closed once GetEvents has finished its job. In order to get the
// events for one stage, n should be the node associated with a Stage.
func getEvents(n *html.Node, e chan<- Event) {
	getEventsRecursive(n, e)
	close(e)
}

// getEventsRecursive is used by GetEvents (and by itself) to walk the running
// order recursively starting at n. Any Event found is published via e.
func getEventsRecursive(n *html.Node, e chan<- Event) {
	if n.Type == html.ElementNode && hasAttributeValue(n.Attr, "class", "band_lineup") {
		nn := node{n}
		name := nn.firstNonEmptyChild().nextNonEmptySibling().FirstChild.Data
		e <- Event{strings.Title(name), n}
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getEventsRecursive(c, e)
	}
}

// A RunningOrder contains the Days of the event.
type RunningOrder struct {
	Days []Day
}

// ParseRunningOrder parses the HTML running order in r and returns a fully
// populated RunningOrder.
func ParseRunningOrder(r io.Reader) (RunningOrder, error) {
	ro := RunningOrder{}

	n, err := html.Parse(r)
	if err != nil {
		return ro, err
	}

	ds := make(chan Day)
	go getDays(n, ds)

	for d := range ds {
		ss := make(chan Stage)
		go getStages(d.node, ss)
		for s := range ss {
			es := make(chan Event)
			go getEvents(s.node, es)
			for e := range es {
				s.Events = append(s.Events, e)
			}

			d.Stages = append(d.Stages, s)
		}

		ro.Days = append(ro.Days, d)
	}

	return ro, nil
}
