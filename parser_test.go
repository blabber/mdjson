// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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
	_, err := getDays(2016, failRootNode)
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
	_, err := getEvents(failRootNode, time.Date(2017, 07, 23, 0, 0, 0, 0, timezone))
	if err == nil {
		t.Error("getEvents(nil) returns no error")
	}
}

type testDay struct {
	Label      string
	TimeStamps *TimeStamps
}

func TestGetDays(t *testing.T) {
	year := 2017

	expected := []testDay{
		{"Saturday 22.07.", &TimeStamps{1500674400, 1500760800}},
		{"Tuesday 25.07.", &TimeStamps{1500933600, 1501020000}},
		{"Wednesday 26.07.", &TimeStamps{1501020000, 1501106400}},
	}

	is, err := getDays(year, sampleRootNode)
	if err != nil {
		t.Fatalf("getDays returned an unexpected error: %v", err)
	}

	for i, isDay := range is {
		if i >= len(expected) {
			break
		}

		if isDay.Label != expected[i].Label {
			t.Errorf("Unexpected label for day %d; is \"%s\"; expected \"%s\"",
				i, isDay.Label, expected[i].Label)
		}

		t.Run(fmt.Sprintf("check timestamps for day %d", i),
			compareTimeStampsPointers(isDay.TimeStamps, expected[i].TimeStamps))
	}

	if len(is) != len(expected) {
		t.Errorf("unexpected number of days found; is %d; expected %d", len(is), len(expected))
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
	Time       string
	TimeStamps *TimeStamps
	Label      string
	URL        string
}

func TestGetEvents(t *testing.T) {
	day := time.Date(2017, 07, 22, 0, 0, 0, 0, timezone)

	expected := []testEvent{
		{
			"-",
			nil,
			strings.Title("Tytus"),
			"http://www.metaldays.net/b613/tytus",
		},
		{
			"-",
			nil,
			strings.Title("Turbowarrior of steel"),
			"http://www.metaldays.net/b612/turbowarrior-of-steel",
		},
		{
			"22:30 - 00:00",
			&TimeStamps{1500755400, 1500760800},
			strings.Title("Amon Amarth"),
			"http://www.metaldays.net/b526/amon-amarth",
		},
		{
			"20:45 - 22:00",
			&TimeStamps{1500749100, 1500753600},
			strings.Title("Katatonia"),
			"http://www.metaldays.net/b531/katatonia",
		},
		{
			"00:10 - 01:20",
			&TimeStamps{1500761400, 1500765600},
			strings.Title("Kadavar"),
			"http://www.metaldays.net/b539/kadavar",
		},
		{
			"22:30 - 00:00",
			&TimeStamps{1500755400, 1500760800},
			strings.Title("Doro"),
			"http://www.metaldays.net/b529/doro",
		},
	}

	is, err := getEvents(sampleRootNode, day)
	if err != nil {
		t.Fatalf("getEvents returned an unexpected error: %v", err)
	}

	for i, isEvent := range is {
		if i >= len(expected) {
			break
		}

		if isEvent.Time != expected[i].Time {
			t.Errorf("Unexpected time for event %d; is \"%s\"; expected \"%s\"",
				i, isEvent.Time, expected[i].Time)
		}

		t.Run(fmt.Sprintf("check timestamps for event %d", i),
			compareTimeStampsPointers(isEvent.TimeStamps, expected[i].TimeStamps))

		if isEvent.Label != expected[i].Label {
			t.Errorf("Unexpected label for event %d; is \"%s\"; expected \"%s\"",
				i, isEvent.Label, expected[i].Label)
		}

		if isEvent.URL != expected[i].URL {
			t.Errorf("Unexpected URL for event %d; is \"%s\"; expected \"%s\"",
				i, isEvent.URL, expected[i].URL)
		}
	}

	if len(is) != len(expected) {
		t.Errorf("unexpected number of events found; is %d; expected %d",
			len(is), len(expected))
	}
}

func TestParseRunningOrder(t *testing.T) {
	year := 2016

	expected := RunningOrder{
		[]*Day{
			{
				"Saturday 22.07.",
				[]*Stage{
					{
						strings.Title("Newforces stage"),
						[]*Event{
							{
								"-",
								nil,
								strings.Title("Tytus"),
								"http://www.metaldays.net/b613/tytus",
							},
							{
								"-",
								nil,
								strings.Title("Turbowarrior of steel"),
								"http://www.metaldays.net/b612/turbowarrior-of-steel",
							},
						},
						nil,
					},
				},
				&TimeStamps{1469138400, 1469224800},
				nil,
			},
			{
				"Tuesday 25.07.",
				[]*Stage{
					{
						strings.Title("Ian Fraser “Lemmy” Kilmister stage"),
						[]*Event{
							{
								"22:30 - 00:00",
								&TimeStamps{1469478600, 1469484000},
								strings.Title("Amon Amarth"),
								"http://www.metaldays.net/b526/amon-amarth",
							},
							{
								"20:45 - 22:00",
								&TimeStamps{1469472300, 1469476800},
								strings.Title("Katatonia"),
								"http://www.metaldays.net/b531/katatonia",
							},
						},
						nil,
					},
					{
						strings.Title("Boško Bursać Stage"),
						[]*Event{
							{
								"00:10 - 01:20",
								&TimeStamps{1469484600, 1469488800},
								strings.Title("Kadavar"),
								"http://www.metaldays.net/b539/kadavar",
							},
						},
						nil,
					},
				},
				&TimeStamps{1469397600, 1469484000},
				nil,
			},
			{
				"Wednesday 26.07.",
				[]*Stage{
					{
						strings.Title("Ian Fraser “Lemmy” Kilmister stage"),
						[]*Event{
							{
								"22:30 - 00:00",
								&TimeStamps{1469565000, 1469570400},
								strings.Title("Doro"),
								"http://www.metaldays.net/b529/doro",
							},
						},
						nil,
					},
				},
				&TimeStamps{1469484000, 1469570400},
				nil,
			},
		},
	}

	f, err := os.Open("./testdata/sample.html")
	if err != nil {
		panic(err)
	}

	ro, err := ParseRunningOrder(year, f)
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
			t.Errorf("unexpected label for day %d; is \"%s\"; expected \"%s\"",
				d, day.Label, expected.Days[d].Label)
			continue
		}

		t.Run(fmt.Sprintf("check timestamps for day %d", d),
			compareTimeStampsPointers(day.TimeStamps, expected.Days[d].TimeStamps))

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

				t.Run(fmt.Sprintf("check timestamps for event %d %d %d", d, s, e),
					compareTimeStampsPointers(event.TimeStamps,
						expected.Days[d].Stages[s].Events[e].TimeStamps))

				if event.URL != expected.Days[d].Stages[s].Events[e].URL {
					t.Errorf("unexpected url for event %d; is \"%s\"; expected \"%s\"",
						e, event.URL,
						expected.Days[d].Stages[s].Events[e].URL)
				}
			}
		}
	}
}
