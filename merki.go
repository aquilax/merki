package main

import (
	"encoding/csv"
	"github.com/joliv/spark"

	"os"
)

type Merki struct{}

func NewMerki() *Merki {
	return &Merki{}
}

func (m *Merki) AddRecord(fileName string, record *Record) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = delimiter
	if err := w.Write(record.getStrings(false)); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func (m *Merki) DrawSparkline(fileName, measure string) (string, error) {
	var values []float64
	parser := NewParser(string(delimiter))
	go parser.ParseFile(getFileName(fileName))
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
	if err != nil {
		return "", err
	}
	sparkline := spark.Line(values)
	return sparkline, nil
}
