package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rofleksey/dredge/internal/app"
)

//go:embed banner.txt
var banner string

func main() {
	defer func() {
		if r := recover(); r != nil {
			sentry.CurrentHub().Recover(r)
			sentry.Flush(2 * time.Second)
			// Intentional: re-panic so the process still exits non-zero after Sentry capture.
			panic(r)
		}
	}()

	fmt.Fprintln(os.Stderr, banner)
	app.New().Run()
}
