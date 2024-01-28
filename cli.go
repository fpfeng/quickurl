package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const Version = "0.0.6"
const DefaultPort = 5731

type CLIParseResult struct {
	Port         int
	Files        []string
	Verbose      bool
	PublicIPOnly bool
}

func update(version string) error {
	log.Debug("start update")
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("fpfeng/quickurl"))
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		fmt.Printf("current version (%s) is the latest", version)
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}
	fmt.Printf("successfully updated to version %s\n", latest.Version())
	return nil
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
				Usage:     "serving `filepath`, -s /path/to/file1 -s /path/to/file2",
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
			&cli.BoolFlag{
				Name:  "update",
				Usage: "update to latest release version",
				Value: false,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.Bool("v") {
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
				log.SetOutput(os.Stdout)
			}

			if cCtx.Bool("help") {
				cli.ShowAppHelp(cCtx)
				return nil
			}

			if cCtx.Bool("update") {
				if err := update(Version); err != nil {
					log.Fatalf("update failed: %v", err)
				}
				return nil
			}

			rawArgs := cCtx.Args().Slice()
			log.Debugf("args: %v", rawArgs)
			servingArgfiles := cCtx.StringSlice("s")
			simpleModeFiles := make([]string, 0)
			for _, arg := range rawArgs { // executed as quickurl file1 file2
				if _, err := os.Stat(arg); err != nil {
					log.Fatal(err)
				}
				simpleModeFiles = append(simpleModeFiles, arg)
			}
			if len(simpleModeFiles) > 0 {
				result.Files = append(result.Files, simpleModeFiles...)
			} else if len(servingArgfiles) > 0 {
				result.Files = append(result.Files, servingArgfiles...)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	return &result
}
