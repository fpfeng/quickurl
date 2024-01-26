package main

import (
	"fmt"
	"os"
	"slices"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const Version = "0.0.5"
const DefaultPort = 5731

type CLIParseResult struct {
	Port         int
	Files        []string
	Verbose      bool
	PublicIPOnly bool
}

func StartCLI(cliArgs []string) *CLIParseResult {
	cli.AppHelpTemplate = `
 NAME:
	{{.Name}} - {{.Usage}}

 USAGE:
	quickurl /path/to/file1 /path/to/file2
	quickurl -s /path/to/file1 -s /path/to/file2 -p 8080
{{if .Commands}}
 OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}`
	result := CLIParseResult{Port: 0, Files: make([]string, 0)}
	app := &cli.App{
		Name:  fmt.Sprintf("QuickURL %v", Version),
		Usage: "Sharing file with http instantly",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "p",
				Usage:       "listening `port number`",
				Value:       DefaultPort,
				Destination: &result.Port,
			},
			&cli.StringSliceFlag{
				Name:      "s",
				Usage:     "serving filepath, -s /path/to/file1 -s /path/to/file2",
				TakesFile: true,
			},
			&cli.BoolFlag{
				Name:        "v",
				Usage:       "verbose log",
				Destination: &result.Verbose,
				Value:       false,
			},
			&cli.BoolFlag{
				Name:        "public-ip",
				Usage:       "print out public ip from external API only",
				Destination: &result.PublicIPOnly,
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

			servingArgfiles := cCtx.StringSlice("s")
			simpleModeFiles := make([]string, 0)
			for _, arg := range rawArgs {
				log.Debug(arg)
				if _, err := os.Stat(arg); err != nil {
					if string(arg[0]) != "-" && len(arg) != 2 {
						log.Fatal(err)
					}
				} else {
					simpleModeFiles = append(simpleModeFiles, arg)
				}
			}
			if len(simpleModeFiles) > 0 {
				result.Files = append(result.Files, simpleModeFiles...)
			} else if len(servingArgfiles) > 0 {
				result.Files = append(result.Files, servingArgfiles...)
			} else {
				cli.ShowAppHelp(cCtx)
				return nil
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	return &result
}
