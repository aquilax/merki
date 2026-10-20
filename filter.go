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

type FilterFunc func(r *Record) (bool, error)

type filter struct {
	w        *csv.Writer
	measures map[string]struct{}
	gi       GroupingInterval
	gt       GroupingType
	a        *accumulator
	ff       []FilterFunc
}

func newFilter(w *csv.Writer, measures []string, gi GroupingInterval, gt GroupingType, ff []FilterFunc) *filter {
	selected := make(map[string]struct{}, len(measures))
	for _, measure := range measures {
		selected[measure] = struct{}{}
	}
	a := make(accumulator)
	return &filter{w, selected, gi, gt, &a, ff}
}

func (f *filter) Add(r *Record) error {
	// apply filters first
	for _, ffunc := range f.ff {
		include, err := ffunc(r)
		if !include {
			return err
		}
	}

	key := ""
	if _, ok := f.measures[r.Measurement]; ok {
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
		f.a.add(accumulatorKey{period: key, measurement: r.Measurement}, r)
	}
	return nil
}

func (f *filter) print() error {
	return f.a.print(f.w, f.gt)
}
