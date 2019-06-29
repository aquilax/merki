package main

import (
	"encoding/csv"
	"fmt"
)

type GroupingInterval int
type GroupingType int
type RoundType int

const (
	intervalNone GroupingInterval = iota
	intervalHourly
	intervalDaily
	intervalWeekly
	intervalTotal

	typeFirst GroupingType = iota
	typeAverage
	typeMax
	typeMin
	typeSum

	roundSeconds RoundType = iota
	roundMinutes
	roundHours
	roundDays
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
		case intervalHourly:
			key = r.Date.Format("2006-01-02 15")
		case intervalDaily:
			key = r.Date.Format("2006-01-02")
		case intervalWeekly:
			year, week := r.Date.ISOWeek()
			key = fmt.Sprintf("%02d-%02d", year, week)
		case intervalTotal:
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
