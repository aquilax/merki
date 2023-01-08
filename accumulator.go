package merki

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
)

// accumulator collects records by string key
type accumulator map[string][]*Record

// Add adds record to the accumulator
func (a *accumulator) add(key string, r *Record) {
	(*a)[key] = append((*a)[key], r)
}

// Print accumulator to the writer
func (a *accumulator) print(w *csv.Writer, gt GroupingType) error {
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

func (a *accumulator) calc(records []*Record, gt GroupingType) float64 {
	var values []float64
	for _, r := range records {
		values = append(values, r.Value)
	}
	switch gt {
	case TypeMax:
		return max(values)
	case TypeMin:
		return min(values)
	case TypeSum:
		return sum(values)
	case TypeAverage:
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
