package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/cli"
)

const (
	fileName = "health.log"

	delimiter = '\t'
)

type Record struct {
	Date        time.Time
	Measurement string
	Value       float64
	Name        string
	Description string
}

func (r *Record) getStrings() []string {
	return []string{
		r.Date.Format("2006-01-02 15:04:05"),
		r.Measurement,
		fmt.Sprintf("%.3f", r.Value),
		r.Name,
		r.Description,
	}
}

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add measurement to file",
			Action: func(c *cli.Context) {
				measurement := c.Args().First()
				value := c.Args().Get(1)
				fValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					panic(err)
				}
				record := &Record{
					time.Now(),
					measurement,
					fValue,
					c.Args().Get(2),
					c.Args().Get(3),
				}
				f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					panic(err)
				}

				defer f.Close()

				w := csv.NewWriter(f)
				w.Comma = delimiter
				if err := w.Write(record.getStrings()); err != nil {
					panic(err)
				}
				w.Flush()
				if err := w.Error(); err != nil {
					log.Fatal(err)
				}
			},
		},
	}

	app.Run(os.Args)
}
