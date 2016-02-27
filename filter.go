package main

import "encoding/csv"

type GroupingInterval int
type GroupingType int

const (
	intervalNone GroupingInterval = iota
	intervalHourly
	intervalDaily
	intervalWeekly

	typeAverage GroupingType = iota
	typeMax
	typeMin
)

type Filter struct {
	w       *csv.Writer
	measure string
	gi      GroupingInterval
	gt      GroupingType
}

func NewFilter(w *csv.Writer, measure string) *Filter {
	return &Filter{w, measure, intervalNone, typeAverage}
}

func (f *Filter) Add(r *Record) error {
	if r.Measurement == f.measure {
		if err := f.w.Write(r.getStrings()); err != nil {
			return err
		}
	}
	return nil
}
