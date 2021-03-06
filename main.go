package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/urfave/cli.v1"
)

const (
	appVersion      = "0.0.8"
	defaultFileName = "health.log"
	delimiter       = '\t'
)

func main() {
	var fileName string

	merki := NewMerki(delimiter)
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
				record, err := NewRecord(time.Now(), args.Get(0), args.Get(1), args.Get(2), args.Get(3))
				if err != nil {
					panic(err)
				}
				return merki.AddRecord(fileName, record)
			},
		},
		{
			Name:    "sparkline",
			Aliases: []string{"spark"},
			Usage:   "Draw sparkline graph for a measure",
			Action: func(c *cli.Context) error {
				sparkline, err := merki.DrawSparkline(fileName, c.Args().First())
				if err == nil {
					fmt.Println(sparkline)
				}
				return err
			},
		},
		{
			Name:    "asciigraph",
			Aliases: []string{"graph"},
			Usage:   "Draw ascii graph for a measure",
			Action: func(c *cli.Context) error {
				graph, err := merki.DrawGraph(fileName, c.Args().First())
				if err == nil {
					fmt.Println(graph)
				}
				return err
			},
		},
		{
			Name:    "measurements",
			Aliases: []string{"m"},
			Usage:   "Return list of all used measurements",
			Action: func(c *cli.Context) error {
				return merki.Measurements(fileName)
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
					Name:  "total, t",
					Usage: "Group values by all time",
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
				gi := intervalNone
				gt := typeAverage
				if c.Bool("hourly") {
					gi = intervalHourly
				}
				if c.Bool("daily") {
					gi = intervalDaily
				}
				if c.Bool("weekly") {
					gi = intervalWeekly
				}
				if c.Bool("total") {
					gi = intervalTotal
				}
				if c.Bool("average") {
					gt = typeAverage
				}
				if c.Bool("max") {
					gt = typeMax
				}
				if c.Bool("min") {
					gt = typeMin
				}
				if c.Bool("sum") {
					gt = typeSum
				}
				return merki.Filter(fileName, measure, gi, gt)
			},
		},
		{
			Name:    "interval",
			Aliases: []string{"i"},
			Usage:   "Shows the interval between two measurement events",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "minutes, m",
					Usage: "Group values by hour",
				},
				cli.BoolFlag{
					Name:  "hours, r",
					Usage: "Group values by hour",
				},
				cli.BoolFlag{
					Name:  "days, d",
					Usage: "Group values by day",
				},
			},
			Action: func(c *cli.Context) error {
				measure := c.Args().First()
				r := roundSeconds
				if c.Bool("minutes") {
					r = roundMinutes
				}
				if c.Bool("hours") {
					r = roundHours
				}
				if c.Bool("days") {
					r = roundDays
				}
				return merki.Interval(fileName, measure, r)
			},
		},
		{
			Name:    "latest",
			Aliases: []string{"l"},
			Usage:   "Show the latest values for all measurements",
			Action: func(c *cli.Context) error {
				return merki.Latest(fileName)
			},
		},
	}

	app.Run(os.Args)
}
