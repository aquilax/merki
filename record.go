package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
)

const (
	formatDate  = "2006-01-02 15:04:05"
	tzFormat    = " -0700 MST"
	formatFloat = "%.3f"
)

type Record struct {
	Date        time.Time
	Measurement string
	Value       float64
	Name        string
	Description string
}

func getS(s []string, l, n int) string {
	if n < l {
		return s[n]
	}
	return ""
}

func NewRecord(timestamp time.Time, measurement, value, name, description string) (*Record, error) {
	fValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	return &Record{
		timestamp,
		measurement,
		fValue,
		name,
		description,
	}, nil
}

func NewRecordFromStrings(s []string) (*Record, error) {
	l := len(s)
	timestamp, err := time.Parse(formatDate+tzFormat, getS(s, l, 0)+time.Now().Format(tzFormat))
	if err != nil {
		return nil, err
	}
	return NewRecord(timestamp, getS(s, l, 1), getS(s, l, 2), getS(s, l, 3), getS(s, l, 4))
}

func (r *Record) getStrings(addRelative bool) []string {
	result := []string{
		r.Date.Format(formatDate),
	}
	if addRelative {
		result = append(result, humanize.RelTime(r.Date, time.Now(), "ago", "later"))
	}
	result = append(result, r.Measurement, fmt.Sprintf(formatFloat, r.Value))
	if r.Name != "" {
		result = append(result, r.Name)
	}
	if r.Description != "" {
		result = append(result, r.Description)
	}
	return result
}
