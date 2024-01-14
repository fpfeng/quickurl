package main

import (
	"os"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const DefaultPort = 5731

type CLIParseResult struct {
	Port    int
	Files   []string
	Verbose bool
}

func StartCLI(cliArgs []string) *CLIParseResult {
	result := CLIParseResult{Port: 0, Files: make([]string, 0)}
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "p",
				Usage:       "listening port",
				Value:       DefaultPort,
				Destination: &result.Port,
			},
			&cli.StringSliceFlag{
				Name:  "s",
				Usage: "serving filepath",
			},
			&cli.BoolFlag{
				Name:        "v",
				Usage:       "verbose log",
				Destination: &result.Verbose,
				Value:       false,
			},
		},
		Action: func(cCtx *cli.Context) error {
			rawArgs := cCtx.Args().Slice()

			if slices.Contains(rawArgs, "-v") {
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
				log.SetOutput(os.Stdout)
			}
			log.Debugf("args: %v", rawArgs)

			simpleModeFiles := make([]string, 0)
			for _, arg := range rawArgs {
				if !strings.Contains(arg, "-") { // is flag?
					if _, err := os.Stat(arg); err != nil {
						log.Fatal(err)
					}
					simpleModeFiles = append(simpleModeFiles, arg)
				}
			}
			if len(simpleModeFiles) > 0 {
				result.Files = append(result.Files, simpleModeFiles...)
			} else {
				result.Files = append(result.Files, cCtx.StringSlice("s")...)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	return &result
}
