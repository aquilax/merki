package main

import (
	"encoding/csv"
	"log"
	"os"
	"sort"

	"gopkg.in/urfave/cli.v1"
)

const (
	appVersion      = "0.0.1"
	defaultFileName = "health.log"
	delimiter       = '\t'
)

func getFileName(fileName string) string {
	if fileName == "" {
		return defaultFileName
	}
	return fileName
}

func main() {
	var fileName string

	merki := NewMerki()
	app := cli.NewApp()
	app.Name = "merki"
	app.Usage = "Command line personal health tracker"
	app.Version = appVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Value:       defaultFileName,
			Usage:       "Log file path",
			EnvVar:      "MERKI_FILE",
			Destination: &fileName,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add measurement value to the file",
			Action: func(c *cli.Context) error {
				args := c.Args()
				record, err := NewRecordFromArgs(args.Get(0), args.Get(1), args.Get(2), args.Get(3))
				if err != nil {
					panic(err)
				}
				return merki.AddRecord(getFileName(fileName), record)
			},
		},
		{
			Name:    "sparkline",
			Aliases: []string{"spark"},
			Usage:   "Draw sparkline graph for a measure",
			Action: func(c *cli.Context) error {
				sparkline, err := merki.DrawSparkline(getFileName(fileName), c.Args().First())
				if err == nil {
					println(sparkline)
				}
				return err
			},
		},
		{
			Name:    "measurements",
			Aliases: []string{"m"},
			Usage:   "Return list of all used measurements",
			Action: func(c *cli.Context) error {
				measures := make(map[string]bool)
				parser := NewParser(string(delimiter))
				go parser.ParseFile(getFileName(fileName))
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
					panic(err)
				}
				for name := range measures {
					println(name)
				}
				return nil
			},
		},
		{
			Name:    "filter",
			Aliases: []string{"f"},
			Usage:   "Filter records for single measurement",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "hourly, r",
					Usage: "Group values by hour",
				},
				cli.BoolFlag{
					Name:  "daily, d",
					Usage: "Group values by day",
				},
				cli.BoolFlag{
					Name:  "weekly, w",
					Usage: "Group values by week",
				},
				cli.BoolFlag{
					Name:  "average, a",
					Usage: "Average values in the group",
				},
				cli.BoolFlag{
					Name:  "max, x",
					Usage: "Max values in the group",
				},
				cli.BoolFlag{
					Name:  "min, n",
					Usage: "Min values in the group",
				},
				cli.BoolFlag{
					Name:  "sum, s",
					Usage: "Sum values in the group",
				},
			},
			Action: func(c *cli.Context) error {
				measure := c.Args().First()
				w := csv.NewWriter(os.Stdout)
				w.Comma = delimiter
				filter := NewFilter(w, measure)
				if c.Bool("hourly") {
					filter.gi = intervalHourly
				}
				if c.Bool("daily") {
					filter.gi = intervalDaily
				}
				if c.Bool("weekly") {
					filter.gi = intervalWeekly
				}
				if c.Bool("average") {
					filter.gt = typeAverage
				}
				if c.Bool("max") {
					filter.gt = typeMax
				}
				if c.Bool("min") {
					filter.gt = typeMin
				}
				if c.Bool("sum") {
					filter.gt = typeSum
				}
				parser := NewParser(string(delimiter))
				go parser.ParseFile(getFileName(fileName))
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
					panic(err)
				}

				err = filter.Print()
				if err != nil {
					panic(err)
				}

				w.Flush()
				if err := w.Error(); err != nil {
					log.Fatal(err)
				}
				return nil
			},
		},
		{
			Name:    "latest",
			Aliases: []string{"l"},
			Usage:   "Show the latest values for all measurements",
			Action: func(c *cli.Context) error {
				w := csv.NewWriter(os.Stdout)
				w.Comma = delimiter
				parser := NewParser(string(delimiter))
				list := make(map[string]*Record)
				var ss sort.StringSlice
				go parser.ParseFile(getFileName(fileName))
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
					panic(err)
				}

				ss.Sort()
				for _, key := range ss {
					r, _ := list[key]
					if err := w.Write(r.getStrings(true)); err != nil {
						log.Fatal(err)
					}
				}

				w.Flush()
				if err := w.Error(); err != nil {
					log.Fatal(err)
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
