package merki

import (
	"encoding/csv"
	"fmt"
)

type GroupingInterval int
type GroupingType int
type RoundType int

const (
	IntervalNone GroupingInterval = iota
	IntervalHourly
	IntervalDaily
	IntervalWeekly
	IntervalTotal

	TypeFirst GroupingType = iota
	TypeAverage
	TypeMax
	TypeMin
	TypeSum

	RoundSeconds RoundType = iota
	RoundMinutes
	RoundHours
	RoundDays
)

type filter struct {
	w       *csv.Writer
	measure string
	gi      GroupingInterval
	gt      GroupingType
	a       *accumulator
}

func newFilter(w *csv.Writer, measure string, gi GroupingInterval, gt GroupingType) *filter {
	a := make(accumulator)
	return &filter{w, measure, gi, gt, &a}
}

func (f *filter) Add(r *Record) error {
	key := ""
	if r.Measurement == f.measure {
		switch f.gi {
		case IntervalHourly:
			key = r.Date.Format("2006-01-02 15")
		case IntervalDaily:
			key = r.Date.Format("2006-01-02")
		case IntervalWeekly:
			year, week := r.Date.ISOWeek()
			key = fmt.Sprintf("%02d-%02d", year, week)
		case IntervalTotal:
			key = "total"
		}
		if key == "" {
			if err := f.w.Write(r.getStrings(false)); err != nil {
				return err
			}
			return nil
		}
		f.a.add(key, r)
	}
	return nil
}

func (f *filter) print() error {
	return f.a.print(f.w, f.gt)
}
