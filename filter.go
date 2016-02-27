package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
)

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
	typeSum
)

type Accumulator map[string][]*Record

type Filter struct {
	w       *csv.Writer
	measure string
	gi      GroupingInterval
	gt      GroupingType
	a       *Accumulator
}

func NewFilter(w *csv.Writer, measure string) *Filter {
	a := make(Accumulator)
	return &Filter{w, measure, intervalNone, typeAverage, &a}
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
		}
		if key == "" {
			if err := f.w.Write(r.getStrings()); err != nil {
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

func (a *Accumulator) Add(key string, r *Record) {
	(*a)[key] = append((*a)[key], r)
}

func (a *Accumulator) Print(w *csv.Writer, gt GroupingType) error {
	var ss sort.StringSlice
	for key, _ := range *a {
		ss = append(ss, key)
	}
	ss.Sort()
	for _, key := range ss {
		records, _ := (*a)[key]
		s := []string{
			key,
			records[0].Measurement,
			fmt.Sprintf(formatFloat, a.calc(records, gt)),
		}
		if err := w.Write(s); err != nil {
			return err
		}
	}
	return nil
}

func (a *Accumulator) calc(records []*Record, gt GroupingType) float64 {
	var values []float64
	for _, r := range records {
		values = append(values, r.Value)
	}
	switch gt {
	case typeMax:
		return max(values)
	case typeMin:
		return min(values)
	case typeSum:
		return sum(values)
	}
	// Default
	return average(values)
}

func sum(v []float64) float64 {
	sum := 0.0
	for _, val := range v {
		sum += val
	}
	return sum
}

func average(v []float64) float64 {
	return sum(v) / float64(len(v))
}

func max(v []float64) float64 {
	res := math.SmallestNonzeroFloat64
	for _, val := range v {
		res = math.Max(res, val)
	}
	return res
}

func min(v []float64) float64 {
	res := math.MaxFloat64
	for _, val := range v {
		res = math.Min(res, val)
	}
	return res
}
