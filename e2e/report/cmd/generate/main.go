package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/xuxiping/port-scan-mk3/e2e/report"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, _ io.Writer, stderr io.Writer) int {
	var out string
	var total int
	var open int
	var closed int
	var timeout int

	fs := flag.NewFlagSet("generate-report", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&out, "out", "e2e/out", "output directory")
	fs.IntVar(&total, "total", 0, "total targets")
	fs.IntVar(&open, "open", 0, "open targets")
	fs.IntVar(&closed, "closed", 0, "closed targets")
	fs.IntVar(&timeout, "timeout", 0, "timeout targets")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	err := report.Generate(out, report.Summary{Total: total, Open: open, Closed: closed, Timeout: timeout})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
