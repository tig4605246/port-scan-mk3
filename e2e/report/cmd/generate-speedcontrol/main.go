package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/xuxiping/port-scan-mk3/e2e/report"
	speedctrlkit "github.com/xuxiping/port-scan-mk3/internal/testkit/speedcontrol"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, _ io.Writer, stderr io.Writer) int {
	var out string

	fs := flag.NewFlagSet("generate-speedcontrol-report", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&out, "out", "e2e/out/speedcontrol", "output directory")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	runs, err := speedctrlkit.RunScenarioMatrix()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := report.GenerateSpeedControlReport(out, runs); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
