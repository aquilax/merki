package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
)

const (
	formatDate  = "2006-01-02 15:04:05"
	formatFloat = "%.3f"
)

type Record struct {
	Date        time.Time
	Measurement string
	Value       float64
	Name        string
	Description string
}

func getS(s []string, n int) string {
	l := len(s)
	if n < l {
		return s[n]
	}
	return ""
}

func NewRecordFromStrings(s []string) (*Record, error) {
	v, err := strconv.ParseFloat(getS(s, 2), 64)
	if err != nil {
		return nil, err
	}
	date, err := time.Parse(formatDate, getS(s, 0))
	if err != nil {
		return nil, err
	}
	return &Record{
		date,
		getS(s, 1),
		v,
		getS(s, 3),
		getS(s, 4),
	}, nil
}

func NewRecordFromArgs(args cli.Args) (*Record, error) {
	measurement := args.First()
	fValue, err := strconv.ParseFloat(args.Get(1), 64)
	if err != nil {
		return nil, err
	}
	return &Record{
		time.Now(),
		measurement,
		fValue,
		args.Get(2),
		args.Get(3),
	}, nil
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
