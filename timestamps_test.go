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

func TestAddTimeStampsToDays(t *testing.T) {
	ts := []struct {
		year     int
		day      *Day
		expected *TimeStamps
	}{
		{2017, &Day{"Saturday 22.07.", nil, nil, nil}, &TimeStamps{1500674400, 1500760800}},
		{2016, &Day{"Wednesday 15.06.", nil, nil, nil}, &TimeStamps{1465941600, 1466028000}},
	}

	for _, test := range ts {
		t.Run(fmt.Sprintf("%s %d", test.day.Label, test.year), func(t *testing.T) {
			err := addTimeStampsToDay(test.year, test.day)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			is := test.day.TimeStamps
			t.Run("check timestamps", compareTimeStampsPointers(is, test.expected))
		})
	}
}

func TestAddTimeStampsToEvent(t *testing.T) {
	ts := []struct {
		event    *Event
		day      time.Time
		expected *TimeStamps
	}{
		{&Event{" - ", nil, "", ""}, time.Date(2017, 7, 22, 0, 0, 0, 0, timezone), nil},
		{&Event{" - ", nil, "", ""}, time.Date(2016, 6, 15, 0, 0, 0, 0, timezone), nil},
		{&Event{"20:30 - 21:15", nil, "", ""}, time.Date(2017, 7, 22, 0, 0, 0, 0, timezone), &TimeStamps{1500748200, 1500750900}},
		{&Event{"23:15 - 00:30", nil, "", ""}, time.Date(2016, 6, 15, 0, 0, 0, 0, timezone), &TimeStamps{1466025300, 1466029800}},
		{&Event{"00:30 - 01:15", nil, "", ""}, time.Date(2016, 6, 15, 0, 0, 0, 0, timezone), &TimeStamps{1466029800, 1466032500}},
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			err := addTimeStampsToEvent(test.event, test.day)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			is := test.event.TimeStamps
			t.Run("check timestamps", compareTimeStampsPointers(is, test.expected))
		})
	}
}
