package main

import (
	"encoding/csv"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/joliv/spark"
)

const (
	defaultFileName = "health.log"

	delimiter = '\t'
)

func getFileName(fileName string) string {
	if fileName == "" {
		return defaultFileName
	}
	return fileName
}

func main() {
	var fileName string

	app := cli.NewApp()
	app.Name = "merki"
	app.Usage = "Command line personal health tracker"

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
			Action: func(c *cli.Context) {
				record, err := NewRecordFromArgs(c.Args())
				if err != nil {
					panic(err)
				}
				f, err := os.OpenFile(getFileName(fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
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
		{
			Name:    "sparkline",
			Aliases: []string{"spark"},
			Usage:   "Draw sparkline graph for a measure",
			Action: func(c *cli.Context) {
				measure := c.Args().First()
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
					panic(err)
				}
				sparkline := spark.Line(values)
				println(sparkline)
			},
		},
	}

	app.Run(os.Args)
}