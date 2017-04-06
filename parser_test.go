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

var (
	failRootNode,
	sampleRootNode *html.Node
)

func TestMain(m *testing.M) {
	var err error

	failRootNode, err = htmlParseFile("./testdata/fail.html")
	if err != nil {
		panic(err)
	}

	sampleRootNode, err = htmlParseFile("./testdata/sample.html")
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func htmlParseFile(s string) (*html.Node, error) {
	f, err := os.Open(s)
	if err != nil {
		return nil, err
	}

	n, err := html.Parse(f)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func TestGetDaysEmpty(t *testing.T) {
	_, err := getDays(failRootNode)
	if err == nil {
		t.Error("getDays(nil) returns no error")
	}
}

func TestGetStagesEmpty(t *testing.T) {
	_, err := getStages(failRootNode)
	if err == nil {
		t.Error("getStages(nil) returns no error")
	}
}

func TestGetEventsEmpty(t *testing.T) {
	_, err := getEvents(failRootNode)
	if err == nil {
		t.Error("getEvents(nil) returns no error")
	}
}

func TestGetDays(t *testing.T) {
	expected := []string{
		"Saturday 22.07.",
		"Tuesday 25.07.",
		"Wednesday 26.07.",
	}

	is, err := getDays(sampleRootNode)
	if err != nil {
		t.Fatalf("getDays returned an unexpected error: %v", err)
	}

	for i, isDay := range is {
		if i >= len(expected) {
			break
		}

		if isLabel := isDay.Label; isLabel != expected[i] {
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

	is, err := getStages(sampleRootNode)
	if err != nil {
		t.Fatalf("getStages returned an unexpected error: %v", err)
	}

	for i, isStage := range is {
		if i >= len(expected) {
			break
		}

		if isLabel := isStage.Label; isLabel != expected[i] {
			t.Errorf("Unexpected label for stage %d; is \"%s\"; expected \"%s\"",
				i, isLabel, expected[i])
		}
	}

	if len(is) != len(expected) {
		t.Errorf("unexpected number of stages found; is %d; expected %d",
			len(is), len(expected))
	}
}

type testEvent struct {
	Label string
	URL   string
}

func TestGetEvents(t *testing.T) {
	expected := []testEvent{
		{strings.Title("Tytus"), "http://www.metaldays.net/b613/tytus"},
		{strings.Title("Turbowarrior of steel"), "http://www.metaldays.net/b612/turbowarrior-of-steel"},
		{strings.Title("Amon Amarth"), "http://www.metaldays.net/b526/amon-amarth"},
		{strings.Title("Katatonia"), "http://www.metaldays.net/b531/katatonia"},
		{strings.Title("Kadavar"), "http://www.metaldays.net/b539/kadavar"},
		{strings.Title("Doro"), "http://www.metaldays.net/b529/doro"},
	}

	is, err := getEvents(sampleRootNode)
	if err != nil {
		t.Fatalf("getStages returned an unexpected error: %v", err)
	}

	for i, isEvent := range is {
		if i >= len(expected) {
			break
		}

		isTestEvent := testEvent{isEvent.Label, isEvent.URL}

		if isTestEvent != expected[i] {
			t.Errorf("Unexpected event %d; is \"%v\"; expected \"%v\"",
				i, isEvent, isTestEvent)
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
							{
								strings.Title("Tytus"),
								"http://www.metaldays.net/b613/tytus",
								nil,
							},
							{
								strings.Title("Turbowarrior of steel"),
								"http://www.metaldays.net/b612/turbowarrior-of-steel",
								nil,
							},
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
							{
								strings.Title("Amon Amarth"),
								"http://www.metaldays.net/b526/amon-amarth",
								nil,
							},
							{
								strings.Title("Katatonia"),
								"http://www.metaldays.net/b531/katatonia",
								nil,
							},
						},
						nil,
					},
					{
						strings.Title("Boško Bursać Stage"),
						[]Event{
							{
								strings.Title("Kadavar"),
								"http://www.metaldays.net/b539/kadavar",
								nil,
							},
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
							{
								strings.Title("Doro"),
								"http://www.metaldays.net/b529/doro",
								nil,
							},
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
	if err != nil {
		t.Fatalf("ParseRunningOrder returns unexpected error: %v", err)
	}

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
					t.Errorf("unexpected label for event %d; is \"%s\"; expected \"%s\"",
						e, event.Label,
						expected.Days[d].Stages[s].Events[e].Label)
				}

				if event.URL != expected.Days[d].Stages[s].Events[e].URL {
					t.Errorf("unexpected url for event %d; is \"%s\"; expected \"%s\"",
						e, event.URL,
						expected.Days[d].Stages[s].Events[e].URL)
				}
			}
		}
	}
}
