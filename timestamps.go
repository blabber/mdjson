// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mdjson

import (
	"strings"
	"time"
)

// timezone is the time.Location where the festival happens.
var timezone *time.Location

// init initializes the global context of the timestamp related functions.
func init() {
	var err error
	timezone, err = time.LoadLocation("Europe/Ljubljana")
	if err != nil {
		panic(err)
	}
}

// A TimeStamps contains two unix timestamps, that denote the start and end of
// a time span.
type TimeStamps struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// addTimeStampsToDay generates TimeStamps for the Day d and adds them to d.
// d.Label has to be filled correctly, before calling this function. year will
// be used as the year component of timestamps.
func addTimeStampsToDay(year int, d *Day) error {
	parsed, e := time.ParseInLocation("Monday 02.01.", d.Label, timezone)
	if e != nil {
		return e
	}

	start := time.Date(year, parsed.Month(), parsed.Day(), 0, 0, 0, 0, timezone)
	end := start.AddDate(0, 0, 1)

	d.TimeStamps = &TimeStamps{start.Unix(), end.Unix()}

	return nil
}

// addTimeStampsToEvent generates TimeStamps for Event e and adds them to e.
// The time.Time d that denotes the start for the day of the event will be used
// to generate the timestamps for the event.
// e.Time has to be filled correctly, before calling this function.
// If the time component of a time stamp is smaller than 10:00, it is assumed,
// that the event belongs to the next day.
func addTimeStampsToEvent(e *Event, d time.Time) error {
	if strings.TrimSpace(e.Time) == "-" {
		return nil
	}

	pf := func(s string, d time.Time) (time.Time, error) {
		p, err := time.ParseInLocation("15:04", s, timezone)
		if err != nil {
			return time.Now(), err
		}

		n := time.Date(d.Year(), d.Month(), d.Day(), p.Hour(), p.Minute(), 0, 0, timezone)
		if n.Hour() < 10 {
			n = n.AddDate(0, 0, 1)
		}

		return n, nil
	}

	timeStrings := strings.Split(e.Time, " - ")
	start, err := pf(timeStrings[0], d)
	if err != nil {
		return err
	}

	end, err := pf(timeStrings[1], d)
	if err != nil {
		return err
	}

	e.TimeStamps = &TimeStamps{start.Unix(), end.Unix()}
	return nil
}
