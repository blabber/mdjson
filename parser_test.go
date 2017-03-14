// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var rootNode *html.Node

func TestMain(m *testing.M) {
	f, err := os.Open("./testdata/sample.html")
	if err != nil {
		panic(err)
	}

	rootNode, err = html.Parse(f)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestGetDays(t *testing.T) {
	expected := []string{
		"Saturday 22.07.",
		"Tuesday 25.07.",
		"Wednesday 26.07.",
	}

	is := []string{}
	d := make(chan Day)
	go getDays(rootNode, d)
	for day := range d {
		is = append(is, day.Label)
	}

	for i, isLabel := range is {
		if i >= len(expected) {
			break
		}

		if isLabel != expected[i] {
			t.Errorf("Unexpected label for day %d; is \"%s\"; expected \"%s\"",
				i, isLabel, expected[i])
		}
	}

	if len(is) != len(expected) {
		t.Errorf("unexpected number of days found; is %d; expected %d",
			len(is), len(expected))
	}
}

func TestGetStages(t *testing.T) {
	expected := []string{
		strings.Title("Newforces stage"),
		strings.Title("Ian Fraser “Lemmy” Kilmister stage"),
		strings.Title("Boško Bursać Stage"),
		strings.Title("Ian Fraser “Lemmy” Kilmister stage"),
	}

	is := []string{}
	s := make(chan Stage)
	go getStages(rootNode, s)
	for stage := range s {
		is = append(is, stage.Label)
	}

	for i, isLabel := range is {
		if i >= len(expected) {
			break
		}

		if isLabel != expected[i] {
			t.Errorf("Unexpected label for stage %d; is \"%s\"; expected \"%s\"",
				i, isLabel, expected[i])
		}
	}

	if len(is) != len(expected) {
		t.Errorf("unexpected number of stages found; is %d; expected %d",
			len(is), len(expected))
	}
}

func TestGetEvents(t *testing.T) {
	expected := []string{
		strings.Title("Tytus"),
		strings.Title("Turbowarrior of steel"),
		strings.Title("Amon Amarth"),
		strings.Title("Katatonia"),
		strings.Title("Kadavar"),
		strings.Title("Doro"),
	}

	is := []string{}
	e := make(chan Event)
	go getEvents(rootNode, e)
	for event := range e {
		is = append(is, event.Label)
	}

	for i, isLabel := range is {
		if i >= len(expected) {
			break
		}

		if isLabel != expected[i] {
			t.Errorf("Unexpected label for event %d; is \"%s\"; expected \"%s\"",
				i, isLabel, expected[i])
		}
	}

	if len(is) != len(expected) {
		t.Errorf("unexpected number of events found; is %d; expected %d",
			len(is), len(expected))
	}
}

func TestParseRunningOrder(t *testing.T) {
	expected := RunningOrder{
		[]Day{
			{
				"Saturday 22.07.",
				[]Stage{
					{
						strings.Title("Newforces stage"),
						[]Event{
							{strings.Title("Tytus"), nil},
							{strings.Title("Turbowarrior of steel"), nil},
						},
						nil,
					},
				},
				nil,
			},
			{
				"Tuesday 25.07.",
				[]Stage{
					{
						strings.Title("Ian Fraser “Lemmy” Kilmister stage"),
						[]Event{
							{strings.Title("Amon Amarth"), nil},
							{strings.Title("Katatonia"), nil},
						},
						nil,
					},
					{
						strings.Title("Boško Bursać Stage"),
						[]Event{
							{strings.Title("Kadavar"), nil},
						},
						nil,
					},
				},
				nil,
			},
			{
				"Wednesday 26.07.",
				[]Stage{
					{
						strings.Title("Ian Fraser “Lemmy” Kilmister stage"),
						[]Event{
							{strings.Title("Doro"), nil},
						},
						nil,
					},
				},
				nil,
			},
		},
	}

	f, err := os.Open("./testdata/sample.html")
	if err != nil {
		panic(err)
	}

	ro, err := ParseRunningOrder(f)

	if len(ro.Days) != len(expected.Days) {
		t.Errorf("unexpected number of days; is %d; expected %d",
			len(ro.Days), len(expected.Days))
	}
	for d, day := range ro.Days {
		if d >= len(expected.Days) {
			break
		}

		if day.Label != expected.Days[d].Label {
			t.Errorf("unexpected day %d; is \"%s\"; expected \"%s\"",
				d, day.Label, expected.Days[d].Label)
			continue
		}

		if len(day.Stages) != len(expected.Days[d].Stages) {
			t.Errorf("unexpected number of stages for day %d; is %d; expected %d",
				d, len(day.Stages), len(expected.Days[d].Stages))
		}
		for s, stage := range day.Stages {
			if s >= len(expected.Days[d].Stages) {
				break
			}

			if stage.Label != expected.Days[d].Stages[s].Label {
				t.Errorf("unexpected stage %d; is \"%s\"; expected \"%s\"",
					s, stage.Label, expected.Days[d].Stages[s].Label)
				continue
			}

			if len(stage.Events) != len(expected.Days[d].Stages[s].Events) {
				t.Errorf("unexpected number of events for day %d, stage %d; is %d; expected %d",
					d, s, len(stage.Events), len(expected.Days[d].Stages[s].Events))
			}
			for e, event := range stage.Events {
				if e >= len(expected.Days[d].Stages[s].Events) {
					break
				}

				if event.Label != expected.Days[d].Stages[s].Events[e].Label {
					t.Errorf("unexpected event %d; is \"%s\"; expected \"%s\"",
						e, event.Label,
						expected.Days[d].Stages[s].Events[e].Label)
					continue
				}
			}
		}
	}
}
