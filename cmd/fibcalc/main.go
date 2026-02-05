package main

import (
	"context"
	"os"

	"github.com/agbru/fibcalc/internal/app"
)

func main() {
	if app.HasVersionFlag(os.Args[1:]) {
		app.PrintVersion(os.Stdout)
		return
	}

	application, err := app.New(os.Args, os.Stderr)
	if err != nil {
		if app.IsHelpError(err) {
			os.Exit(0)
		}
		os.Exit(1)
	}

	exitCode := application.Run(context.Background(), os.Stdout)
	os.Exit(exitCode)
}
