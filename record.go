package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codegangsta/cli"
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

func (r *Record) getStrings() []string {
	result := []string{
		r.Date.Format(formatDate),
		r.Measurement,
		fmt.Sprintf(formatFloat, r.Value),
	}
	if r.Name != "" {
		result = append(result, r.Name)
	}
	if r.Description != "" {
		result = append(result, r.Description)
	}
	return result
}
