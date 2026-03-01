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
