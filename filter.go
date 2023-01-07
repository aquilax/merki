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

type Filter struct {
	w       *csv.Writer
	measure string
	gi      GroupingInterval
	gt      GroupingType
	a       *Accumulator
}

func NewFilter(w *csv.Writer, measure string, gi GroupingInterval, gt GroupingType) *Filter {
	a := make(Accumulator)
	return &Filter{w, measure, gi, gt, &a}
}

func (f *Filter) Add(r *Record) error {
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
		f.a.Add(key, r)
	}
	return nil
}

func (f *Filter) Print() error {
	return f.a.Print(f.w, f.gt)
}
