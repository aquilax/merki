package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"
)

type Parser struct {
	delimiter string
	Record    chan *Record
	Error     chan error
	Done      chan bool
}

func NewParser(delimiter string) *Parser {
	return &Parser{
		delimiter,
		make(chan *Record),
		make(chan error),
		make(chan bool),
	}
}

func (p *Parser) ParseFile(fileName string) {
	f, err := os.Open(fileName)
	if err != nil {
		p.Error <- err
		return
	}
	defer f.Close()
	p.ParseStream(f)
}

func (p *Parser) ParseStream(reader io.Reader) {
	lineNumber := 0
	lineScanner := bufio.NewScanner(reader)
	for lineScanner.Scan() {
		lineNumber++
		line := lineScanner.Text()
		segments := strings.Split(line, p.delimiter)
		record, err := NewRecordFromStrings(segments)
		if err != nil {
			p.Error <- err
			continue
		}
		p.Record <- record
	}
	p.Done <- true
}

type ParserCallback func(r *Record, err error) (bool, error)

func ParseFileCallback(fileName string, delimiter rune, cb ParserCallback) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return ParseStreamCallback(f, delimiter, cb)
}

func ParseStreamCallback(reader io.Reader, delimiter rune, cb ParserCallback) error {
	lineNumber := 0
	r := csv.NewReader(reader)
	r.Comma = delimiter
	r.FieldsPerRecord = -1
	for {
		lineNumber++
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if stop, cbErr := cb(nil, err); stop {
				return cbErr
			}
		}
		if stop, cbErr := cb(NewRecordFromStrings(record)); stop {
			return cbErr
		}
	}
	return nil
}
