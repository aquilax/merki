package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
)

// Accumulator collects records by string key
type Accumulator map[string][]*Record

// Add adds record to the accumulator
func (a *Accumulator) Add(key string, r *Record) {
	(*a)[key] = append((*a)[key], r)
}

// Print accumulator to the writer
func (a *Accumulator) Print(w *csv.Writer, gt GroupingType) error {
	var ss sort.StringSlice
	for key := range *a {
		ss = append(ss, key)
	}
	ss.Sort()
	for _, key := range ss {
		records := (*a)[key]
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
	case typeAverage:
		return average(values)
	}
	if len(values) > 0 {
		return first(values)
	}
	return 0
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

func first(v []float64) float64 {
	return v[0]
}
