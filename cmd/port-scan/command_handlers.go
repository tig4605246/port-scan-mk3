package main

import (
	"fmt"
	"io"

	"github.com/xuxiping/port-scan-mk3/pkg/cli"
	"github.com/xuxiping/port-scan-mk3/pkg/config"
)

func handleHelpCommand(stdout io.Writer) int {
	usage(stdout)
	return 0
}

func handleValidateCommand(args []string, stdout, stderr io.Writer) int {
	cfg, err := config.Parse(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	valid, detail := validateInputs(cfg)
	if err := cli.WriteValidation(stdout, cfg.Format, valid, detail); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if !valid {
		return 1
	}
	return 0
}

func handleScanCommand(args []string, stdout, stderr io.Writer) int {
	return runScan(args, stdout, stderr)
}
