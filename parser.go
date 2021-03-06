// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// A Day represents a day.
type Day struct {
	// Label contains a string representation of the date, ideally the date.
	Label string `json:"label"`

	// Stages contains the stages that are active at this day.
	Stages []*Stage `json:"stages"`

	// TimeStamps contains the timestamps for the start and the end of the day.
	TimeStamps *TimeStamps `json:"timestamps"`

	// node is the *html.Node associated with the day. It is intended to be
	// used as input to getStages.
	node *html.Node
}

// getDays walks the running order starting at n and returns a slice of found
// Days. year should be the year the days belong to.
func getDays(year int, n *html.Node) ([]*Day, error) {
	d := make(chan *Day)
	e := make(chan error)
	done := make(chan bool)

	go protect(e, done, func() {
		getDaysRecursive(year, n, d)
	})

	days := []*Day{}
	for {
		select {
		case dd := <-d:
			days = append(days, dd)
		case err := <-e:
			return nil, err
		case <-done:
			return days, nil
		}
	}
}

// getDaysRecursive is used by GetDays (and by itself) to walk the running
// order recursively starting at n. Any Day found is published via d. year
// should be the year the days belong to.
func getDaysRecursive(year int, n *html.Node, d chan<- *Day) {
	if n.Type == html.ElementNode && hasAttributeValue(n.Attr, "class", "lineup_day") {
		nn := newNode(n)
		datenode := nn.firstNonEmptyChild().nextNonEmptySibling().firstNonEmptyChild().firstNonEmptyChild()
		if datenode == nil {
			panic("Unable to parse running order structure (day).")
		}

		// For some reason there is an additional space behind each date
		// separator
		date := strings.Replace(datenode.Data, ". ", ".", -1)
		date = strings.TrimSpace(date)

		day := &Day{date, []*Stage{}, nil, n}

		e := addTimeStampsToDay(year, day)
		if e != nil {
			panic(e)
		}

		d <- day
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getDaysRecursive(year, c, d)
	}
}

// A Stage represents a stage.
type Stage struct {
	// Label contains the name of the stage.
	Label string `json:"label"`

	// Events contains the events that will take on the stage.
	Events []*Event `json:"events"`

	// node is the *html.Node associated with the stage. It is intended to
	// be used as input for getEvents.
	node *html.Node
}

// getStages walks the running order starting at n and returns a slice of found
// Stages. In order to get the stages for one day, n should be the node
// associated with a Day.
func getStages(n *html.Node) ([]*Stage, error) {
	s := make(chan *Stage)
	e := make(chan error)
	done := make(chan bool)

	go protect(e, done, func() {
		getStagesRecursive(n, s)
	})

	stages := []*Stage{}
	for {
		select {
		case ss := <-s:
			stages = append(stages, ss)
		case err := <-e:
			return nil, err
		case <-done:
			return stages, nil
		}
	}
}

// getStagesRecursive is used by GetStages (and by itself) to walk the running
// order recursively starting at n. Any Stage found is published via s.
func getStagesRecursive(n *html.Node, s chan<- *Stage) {
	if n.Type == html.ElementNode && hasAttributeValue(n.Attr, "class", "lineup_stage") {
		nn := newNode(n)
		namenode := nn.firstNonEmptyChild().firstNonEmptyChild().nextNonEmptySibling().firstNonEmptyChild()
		if namenode == nil {
			panic("Unable to parse running order structure (stage).")
		}

		name := strings.TrimSpace(namenode.Data)
		name = strings.Title(name)

		s <- &Stage{name, []*Event{}, n}
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getStagesRecursive(c, s)
	}
}

// A Event represents an event.
type Event struct {
	// Time contains the time of the event
	Time string `json:"time"`

	// TimeStamps contains the timestamps for the start and the end of the event.
	TimeStamps *TimeStamps `json:"timestamps"`

	// Label contains the name of the event, normally the name of a band.
	Label string `json:"label"`

	// URL contains a string representation of an URL that points to
	// additional information about the event.
	URL string `json:"url"`
}

// getEvents walks the running order starting at n and returns a slice of found
// Events.
// The time.Time d that denotes the start for the day of the event will be used
// to generate the timestamps for the event.
// In order to get the events for one stage, n should be the node associated
// with a Stage.
func getEvents(n *html.Node, d time.Time) ([]*Event, error) {
	ev := make(chan *Event)
	e := make(chan error)
	done := make(chan bool)

	go protect(e, done, func() {
		getEventsRecursive(n, d, ev)
	})

	events := []*Event{}
	for {
		select {
		case ee := <-ev:
			events = append(events, ee)
		case err := <-e:
			return nil, err
		case <-done:
			return events, nil
		}
	}
}

// getEventsRecursive is used by GetEvents (and by itself) to walk the running
// order recursively starting at n.
// The time.Time d that denotes the start for the day of the event will be used
// to generate the timestamps for the even.
// Any Event found is published via e.
func getEventsRecursive(n *html.Node, d time.Time, e chan<- *Event) {
	if n.Type == html.ElementNode && hasAttributeValue(n.Attr, "class", "band_lineup") {
		nn := newNode(n)

		timenode := nn.firstNonEmptyChild().nextNonEmptySibling()
		namenode := nn.firstNonEmptyChild().nextNonEmptySibling().nextNonEmptySibling()

		// At the moment the DOM of the Newforces stage is different.
		if timenode == nil || !hasAttributeValue(timenode.Attr, "class", "time") {
			timenode = nn.firstNonEmptyChild()
			namenode = nn.firstNonEmptyChild().nextNonEmptySibling()
		}

		timenode = timenode.firstNonEmptyChild()
		namenode = namenode.firstNonEmptyChild()

		if namenode == nil || timenode == nil {
			panic("Unable to parse running order structure (event).")
		}

		time := strings.TrimSpace(timenode.Data)
		name := strings.TrimSpace(namenode.Data)
		name = strings.ToLower(name)
		name = strings.Title(name)

		url := getAttributeValue(n.Attr, "href")

		event := &Event{time, nil, name, url}
		err := addTimeStampsToEvent(event, d)
		if err != nil {
			panic(err)
		}

		e <- event
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getEventsRecursive(c, d, e)
	}
}

// A RunningOrder contains the Days of the event.
type RunningOrder struct {
	Days []*Day `json:"days"`
}

// ParseRunningOrder parses the HTML running order in r and returns a fully
// populated RunningOrder. year is the year in which the festival takes place;
// generally this should be time.Now().Year().
func ParseRunningOrder(year int, r io.Reader) (*RunningOrder, error) {
	n, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	ro := &RunningOrder{}

	ro.Days, err = getDays(year, n)
	if err != nil {
		return nil, err
	}

	for _, d := range ro.Days {
		d.Stages, err = getStages(d.node)
		if err != nil {
			return nil, err
		}

		dd := time.Unix(d.TimeStamps.Start, 0).In(timezone)
		for _, s := range d.Stages {
			s.Events, err = getEvents(s.node, dd)
			if err != nil {
				return nil, err
			}
		}
	}

	return ro, nil
}
