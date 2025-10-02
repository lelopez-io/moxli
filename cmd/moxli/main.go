package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"
)

var (
	buildVersion = "dev"
	commit       = "none"
	date         = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "moxli",
		Usage:   "Bookmark management CLI",
		Version: buildVersion,
		Commands: []*cli.Command{
			versionCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(c *cli.Context) error {
			fmt.Printf("Version:    %s\n", buildVersion)
			fmt.Printf("Commit:     %s\n", commit)
			fmt.Printf("Build Date: %s\n", date)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}
