package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/aquilax/merki"
	"github.com/urfave/cli/v2"
)

const (
	appVersion      = "0.0.9"
	defaultFileName = "health.log"
	delimiter       = '\t'
)

func main() {
	var fileName string

	output := os.Stdout
	m := merki.New(delimiter, output)
	app := cli.NewApp()
	app.Name = "merki"
	app.Usage = "Command line personal health tracker"
	app.Version = appVersion
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "file",
			Aliases:     []string{"f"},
			Value:       defaultFileName,
			Usage:       "Log file path",
			EnvVars:     []string{"MERKI_FILE"},
			Destination: &fileName,
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add measurement value to the file",
			Action: func(c *cli.Context) error {
				args := c.Args()
				fValue, err := strconv.ParseFloat(args.Get(1), 64)
				if err != nil {
					return err
				}
				record, err := merki.NewRecord(time.Now(), args.Get(0), fValue, args.Get(2), args.Get(3))
				if err != nil {
					return err
				}
				return withWriter(fileName, func(w io.Writer) error {
					return m.AddRecord(w, record)
				})

			},
		},
		{
			Name:    "sparkline",
			Aliases: []string{"spark"},
			Usage:   "Draw sparkline graph for a measure",
			Action: func(c *cli.Context) error {
				return withReader(fileName, func(r io.Reader) error {
					sparkLine, err := m.DrawSparkLine(r, c.Args().First())
					if err != nil {
						return err
					}
					_, err = fmt.Fprintln(output, sparkLine)
					return err
				})
			},
		},
		{
			Name:    "asciigraph",
			Aliases: []string{"graph"},
			Usage:   "Draw ascii graph for a measure",
			Action: func(c *cli.Context) error {
				return withReader(fileName, func(r io.Reader) error {
					graph, err := m.DrawGraph(r, c.Args().First())
					if err == nil {
						return err
					}
					_, err = fmt.Fprintln(output, graph)
					return err
				})
			},
		},
		{
			Name:    "measurements",
			Aliases: []string{"m"},
			Usage:   "Return list of all used measurements",
			Action: func(c *cli.Context) error {
				return withReader(fileName, func(r io.Reader) error {
					return m.Measurements(r)
				})
			},
		},
		{
			Name:    "filter",
			Aliases: []string{"f"},
			Usage:   "Filter records for single measurement",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "hourly",
					Aliases: []string{"r"},
					Usage:   "Group values by hour",
				},
				&cli.BoolFlag{
					Name:    "daily",
					Aliases: []string{"d"},
					Usage:   "Group values by day",
				},
				&cli.BoolFlag{
					Name:    "weekly",
					Aliases: []string{"w"},
					Usage:   "Group values by week",
				},
				&cli.BoolFlag{
					Name:    "total",
					Aliases: []string{"t"},
					Usage:   "Group values by all time",
				},
				&cli.BoolFlag{
					Name:    "average",
					Aliases: []string{"a"},
					Usage:   "Average values in the group",
				},
				&cli.BoolFlag{
					Name:    "max",
					Aliases: []string{"x"},
					Usage:   "Max values in the group",
				},
				&cli.BoolFlag{
					Name:    "min",
					Aliases: []string{"n"},
					Usage:   "Min values in the group",
				},
				&cli.BoolFlag{
					Name:    "sum",
					Aliases: []string{"s"},
					Usage:   "Sum values in the group",
				},
			},
			Action: func(c *cli.Context) error {
				measure := c.Args().First()
				gi := merki.IntervalNone
				gt := merki.TypeAverage
				if c.Bool("hourly") {
					gi = merki.IntervalHourly
				}
				if c.Bool("daily") {
					gi = merki.IntervalDaily
				}
				if c.Bool("weekly") {
					gi = merki.IntervalWeekly
				}
				if c.Bool("total") {
					gi = merki.IntervalTotal
				}
				if c.Bool("average") {
					gt = merki.TypeAverage
				}
				if c.Bool("max") {
					gt = merki.TypeMax
				}
				if c.Bool("min") {
					gt = merki.TypeMin
				}
				if c.Bool("sum") {
					gt = merki.TypeSum
				}
				return withReader(fileName, func(r io.Reader) error {
					return m.Filter(r, measure, gi, gt)
				})
			},
		},
		{
			Name:    "interval",
			Aliases: []string{"i"},
			Usage:   "Shows the interval between two measurement events",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "minutes",
					Aliases: []string{"m"},
					Usage:   "Group values by hour",
				},
				&cli.BoolFlag{
					Name:    "hours",
					Aliases: []string{"r"},
					Usage:   "Group values by hour",
				},
				&cli.BoolFlag{
					Name:    "days",
					Aliases: []string{"d"},
					Usage:   "Group values by day",
				},
			},
			Action: func(c *cli.Context) error {
				measure := c.Args().First()
				round := merki.RoundSeconds
				if c.Bool("minutes") {
					round = merki.RoundMinutes
				}
				if c.Bool("hours") {
					round = merki.RoundHours
				}
				if c.Bool("days") {
					round = merki.RoundDays
				}
				return withReader(fileName, func(r io.Reader) error {
					return m.Interval(r, measure, round)
				})
			},
		},
		{
			Name:    "latest",
			Aliases: []string{"l"},
			Usage:   "Show the latest values for all measurements",
			Action: func(c *cli.Context) error {
				return withReader(fileName, func(r io.Reader) error {
					return m.Latest(r)
				})
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func withReader(fileName string, cb func(r io.Reader) error) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return cb(f)
}

func withWriter(fileName string, cb func(w io.Writer) error) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()
	return cb(f)
}
