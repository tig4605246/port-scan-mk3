package main

import (
	"fmt"
	"io"
	"os"

	"github.com/xuxiping/port-scan-mk3/pkg/cli"
	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/input"
)

func main() {
	os.Exit(runMain(os.Args[1:], os.Stdout, os.Stderr))
}

func runMain(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		usage(stdout)
		return 0
	}

	switch args[0] {
	case "validate":
		cfg, err := config.Parse(args[1:])
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
	case "scan":
		return runScan(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		return 2
	}
}

func runScan(args []string, stdout, stderr io.Writer) int {
	_, err := config.Parse(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	_, _ = fmt.Fprintln(stdout, "scan mode is wired; pipeline implementation pending")
	return 0
}

func validateInputs(cfg config.Config) (bool, string) {
	cidrFile, err := os.Open(cfg.CIDRFile)
	if err != nil {
		return false, fmt.Sprintf("failed to open cidr file: %v", err)
	}
	defer cidrFile.Close()

	if _, err := input.LoadCIDRs(cidrFile); err != nil {
		return false, err.Error()
	}

	portFile, err := os.Open(cfg.PortFile)
	if err != nil {
		return false, fmt.Sprintf("failed to open port file: %v", err)
	}
	defer portFile.Close()

	if _, err := input.LoadPorts(portFile); err != nil {
		return false, err.Error()
	}
	return true, "ok"
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "port-scan scan -cidr-file <file> -port-file <file> [flags]")
	fmt.Fprintln(w, "port-scan validate -cidr-file <file> -port-file <file> [-format human|json]")
	fmt.Fprintln(w, "Flags: -resume -disable-api -pressure-api -pressure-interval -bucket-rate -bucket-capacity -workers -timeout -delay -log-level -format")
}
