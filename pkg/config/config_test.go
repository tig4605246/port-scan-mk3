package config

import (
	"testing"
	"time"
)

func TestParseConfig_WhenRequiredFlagsProvided_UsesDefaultValues(t *testing.T) {
	cfg, err := Parse([]string{"-cidr-file", "cidr.csv", "-port-file", "ports.csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout != 100*time.Millisecond {
		t.Fatalf("timeout mismatch: %v", cfg.Timeout)
	}
	if cfg.Format != "human" {
		t.Fatalf("format mismatch: %s", cfg.Format)
	}
}

func TestParseConfig_WhenPortFileOmitted_StillParses(t *testing.T) {
	cfg, err := Parse([]string{"-cidr-file", "cidr.csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PortFile != "" {
		t.Fatalf("expected empty port file, got %q", cfg.PortFile)
	}
}

func TestParseConfig_WhenCIDRFileMissing_ReturnsError(t *testing.T) {
	_, err := Parse([]string{"-port-file", "ports.csv"})
	if err == nil {
		t.Fatal("expected error for missing cidr-file")
	}
}

func TestParseConfig_WhenFormatIsInvalid_ReturnsError(t *testing.T) {
	_, err := Parse([]string{"-cidr-file", "cidr.csv", "-port-file", "ports.csv", "-format", "xml"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestParseConfig_WhenPressureIntervalProvidedInSeconds_ParsesDuration(t *testing.T) {
	cfg, err := Parse([]string{
		"-cidr-file", "cidr.csv",
		"-port-file", "ports.csv",
		"-pressure-interval", "7",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PressureInterval != 7*time.Second {
		t.Fatalf("expected 7s, got %v", cfg.PressureInterval)
	}
}

func TestParseConfig_WhenCIDRColumnFlagsProvided_SetsColumnNames(t *testing.T) {
	cfg, err := Parse([]string{
		"-cidr-file", "cidr.csv",
		"-port-file", "ports.csv",
		"-cidr-ip-col", "source_ip",
		"-cidr-ip-cidr-col", "source_cidr",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.CIDRIPCol != "source_ip" || cfg.CIDRIPCidrCol != "source_cidr" {
		t.Fatalf("unexpected cols: %#v", cfg)
	}
}

func TestParseConfig_WhenCIDRIPColumnIsEmpty_ReturnsError(t *testing.T) {
	_, err := Parse([]string{
		"-cidr-file", "cidr.csv",
		"-port-file", "ports.csv",
		"-cidr-ip-col", " ",
		"-cidr-ip-cidr-col", "source_cidr",
	})
	if err == nil {
		t.Fatal("expected error for empty cidr-ip-col")
	}
}

func TestParseConfig_WhenOutputAndResumeFlagsProvided_SetsPaths(t *testing.T) {
	cfg, err := Parse([]string{
		"-cidr-file", "cidr.csv",
		"-port-file", "ports.csv",
		"-output", "/tmp/out.csv",
		"-resume", "/tmp/resume_state.json",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.Output != "/tmp/out.csv" {
		t.Fatalf("unexpected output: %s", cfg.Output)
	}
	if cfg.Resume != "/tmp/resume_state.json" {
		t.Fatalf("unexpected resume path: %s", cfg.Resume)
	}
}
