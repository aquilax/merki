package main

import (
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/joliv/spark"

	"os"
)

type Merki struct {
	delimiter rune
}

func NewMerki(delimier rune) *Merki {
	return &Merki{delimiter}
}

func (m *Merki) AddRecord(fileName string, record *Record) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = m.delimiter
	if err := w.Write(record.getStrings(false)); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func getMeasureValues(delimiter, fileName, measure string) ([]float64, error) {
	var values []float64
	parser := NewParser(delimiter)
	go parser.ParseFile(fileName)
	err := func() error {
		for {
			select {
			case record := <-parser.Record:
				if record.Measurement == measure {
					values = append(values, record.Value)
				}
			case err := <-parser.Error:
				return err
			case <-parser.Done:
				return nil
			}
		}
	}()
	return values, err
}

func (m *Merki) DrawGraph(fileName, measure string) (string, error) {
	values, err := getMeasureValues(string(m.delimiter), fileName, measure)
	if err != nil || len(values) == 0 {
		return "", err
	}
	graph := asciigraph.Plot(values, asciigraph.Width(80))
	return graph, err
}

func (m *Merki) DrawSparkline(fileName, measure string) (string, error) {
	values, err := getMeasureValues(string(m.delimiter), fileName, measure)
	if err != nil {
		return "", err
	}
	sparkline := spark.Line(values)
	return sparkline, nil
}

func (m *Merki) Measurements(fileName string) error {
	measures := make(map[string]bool)
	parser := NewParser(string(m.delimiter))
	go parser.ParseFile(fileName)
	err := func() error {
		for {
			select {
			case record := <-parser.Record:
				measures[record.Measurement] = true
			case err := <-parser.Error:
				return err
			case <-parser.Done:
				return nil
			}
		}
	}()
	if err != nil {
		return err
	}
	for name := range measures {
		fmt.Println(name)
	}
	return nil
}

func (m *Merki) Latest(fileName string) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = m.delimiter
	parser := NewParser(string(m.delimiter))
	list := make(map[string]*Record)
	var ss sort.StringSlice
	go parser.ParseFile(fileName)
	err := func() error {
		for {
			select {
			case record := <-parser.Record:
				key := record.Measurement
				val, ok := list[key]
				if !ok {
					list[key] = record
					ss = append(ss, key)
					continue
				}
				if record.Date.After(val.Date) {
					list[key] = record
				}
			case err := <-parser.Error:
				return err
			case <-parser.Done:
				return nil
			}
		}
	}()
	if err != nil {
		return err
	}
	ss.Sort()
	for _, key := range ss {
		r, _ := list[key]
		if err := w.Write(r.getStrings(true)); err != nil {
			return err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func (m *Merki) Filter(fileName, measure string, gi GroupingInterval, gt GroupingType) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = m.delimiter
	filter := NewFilter(w, measure, gi, gt)
	parser := NewParser(string(m.delimiter))
	go parser.ParseFile(fileName)
	err := func() error {
		for {
			select {
			case record := <-parser.Record:
				if err := filter.Add(record); err != nil {
					return err
				}
			case err := <-parser.Error:
				return err
			case <-parser.Done:
				return nil
			}
		}
	}()
	if err != nil {
		return err
	}

	err = filter.Print()
	if err != nil {
		return err
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func formatDuration(d time.Duration, r RoundType) string {
	if r == roundDays {
		return fmt.Sprintf(formatFloat, d.Hours()/24)
	}
	if r == roundHours {
		return fmt.Sprintf(formatFloat, d.Hours())
	}
	if r == roundMinutes {
		return fmt.Sprintf(formatFloat, d.Minutes())
	}
	return fmt.Sprintf("%d", int(d.Seconds()))
}

func (m *Merki) Interval(fileName, measure string, r RoundType) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = m.delimiter
	parser := NewParser(string(m.delimiter))
	go parser.ParseFile(fileName)
	err := func() error {
		var startTime *time.Time
		var duration time.Duration
		var record *Record
		var lastRecord *Record
		for {
			select {
			case record = <-parser.Record:
				if record.Measurement == measure {
					lastRecord = record
					if startTime != nil {
						duration = record.Date.Sub(*startTime)
						err := w.Write([]string{
							record.Date.Format(formatDate),
							measure,
							formatDuration(duration, r),
						})
						if err != nil {
							return err
						}
					}
					startTime = &record.Date
				}
			case err := <-parser.Error:
				return err
			case <-parser.Done:
				if startTime != nil && lastRecord != nil {
					duration = time.Now().Sub(*startTime)
					err := w.Write([]string{
						lastRecord.Date.Format(formatDate),
						measure,
						formatDuration(duration, r),
					})
					if err != nil {
						return err
					}
				}

				return nil
			}
		}
	}()
	if err != nil {
		return err
	}
	w.Flush()
	if err != nil {
		return err
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}
