package merki

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/joliv/spark"
)

type Merki struct {
	delimiter rune
	output    io.Writer
}

func New(delimiter rune, o io.Writer) *Merki {
	return &Merki{delimiter, o}
}

func (m *Merki) AddRecord(w io.Writer, record *Record) error {
	wr := csv.NewWriter(w)
	wr.Comma = m.delimiter
	if err := wr.Write(record.getStrings(false)); err != nil {
		return err
	}
	wr.Flush()
	return wr.Error()
}

func getMeasureValues(r io.Reader, delimiter rune, measure string) ([]float64, error) {
	var values []float64
	err := ParseStreamCallback(r, delimiter, func(record *Record, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		if record.Measurement == measure {
			values = append(values, record.Value)
		}
		return false, nil
	})
	return values, err
}

func (m *Merki) DrawGraph(r io.Reader, measure string) (string, error) {
	values, err := getMeasureValues(r, m.delimiter, measure)
	if err != nil || len(values) == 0 {
		return "", err
	}
	graph := asciigraph.Plot(values, asciigraph.Width(80))
	return graph, err
}

func (m *Merki) DrawSparkLine(r io.Reader, measure string) (string, error) {
	values, err := getMeasureValues(r, m.delimiter, measure)
	if err != nil {
		return "", err
	}
	return spark.Line(values), nil
}

func (m *Merki) Measurements(r io.Reader) error {
	measures := make(map[string]bool)
	err := ParseStreamCallback(r, m.delimiter, func(record *Record, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		measures[record.Measurement] = true
		return false, nil
	})
	if err != nil {
		return err
	}
	for name := range measures {
		if _, err = fmt.Fprintln(m.output, name); err != nil {
			return err
		}
	}
	return err
}

func (m *Merki) Latest(r io.Reader) error {
	w := csv.NewWriter(m.output)
	w.Comma = m.delimiter
	list := make(map[string]*Record)
	var ss sort.StringSlice
	err := ParseStreamCallback(r, m.delimiter, func(record *Record, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		key := record.Measurement
		val, ok := list[key]
		if !ok {
			list[key] = record
			ss = append(ss, key)
		} else if record.Date.After(val.Date) {
			list[key] = record
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	ss.Sort()
	for _, key := range ss {
		r := list[key]
		if err := w.Write(r.getStrings(true)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func (m *Merki) Filter(r io.Reader, measure string, gi GroupingInterval, gt GroupingType) error {
	w := csv.NewWriter(m.output)
	w.Comma = m.delimiter
	filter := NewFilter(w, measure, gi, gt)
	err := ParseStreamCallback(r, m.delimiter, func(record *Record, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		if err = filter.Add(record); err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	if err = filter.Print(); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func formatDuration(d time.Duration, r RoundType) string {
	if r == RoundDays {
		return fmt.Sprintf(formatFloat, d.Hours()/24)
	}
	if r == RoundHours {
		return fmt.Sprintf(formatFloat, d.Hours())
	}
	if r == RoundMinutes {
		return fmt.Sprintf(formatFloat, d.Minutes())
	}
	return fmt.Sprintf("%d", int(d.Seconds()))
}

func (m *Merki) Interval(r io.Reader, measure string, round RoundType) error {
	w := csv.NewWriter(m.output)
	w.Comma = m.delimiter

	var startTime *time.Time
	var duration time.Duration
	var lastRecord *Record

	err := ParseStreamCallback(r, m.delimiter, func(record *Record, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		if record.Measurement == measure {
			lastRecord = record
			if startTime != nil {
				duration = record.Date.Sub(*startTime)
				if err := w.Write([]string{
					record.Date.Format(formatDate),
					measure,
					formatDuration(duration, round),
				}); err != nil {
					return true, err
				}
			}
			startTime = &record.Date
		}

		return false, nil
	})
	if err != nil {
		return err
	}
	if startTime != nil && lastRecord != nil {
		duration = time.Since(*startTime)
		if err = w.Write([]string{
			lastRecord.Date.Format(formatDate),
			measure,
			formatDuration(duration, round),
		}); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
