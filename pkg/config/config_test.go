package config

import (
	"testing"
	"time"
)

func TestParseConfig_Defaults(t *testing.T) {
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

func TestParseConfig_InvalidFormat(t *testing.T) {
	_, err := Parse([]string{"-cidr-file", "cidr.csv", "-port-file", "ports.csv", "-format", "xml"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestParseConfig_PressureIntervalSeconds(t *testing.T) {
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

func TestParseConfig_CIDRColumnFlags(t *testing.T) {
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

func TestParseConfig_CIDRColumnFlags_NonEmpty(t *testing.T) {
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
