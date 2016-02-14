package main

import (
	"bufio"
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
