package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	CIDRFile         string
	PortFile         string
	Output           string
	Timeout          time.Duration
	Delay            time.Duration
	BucketRate       int
	BucketCapacity   int
	Workers          int
	PressureAPI      string
	PressureInterval time.Duration
	DisableAPI       bool
	Resume           string
	LogLevel         string
	Format           string
	CIDRIPCol        string
	CIDRIPCidrCol    string
}

func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("port-scan", flag.ContinueOnError)
	cfg := Config{}
	pressureIntervalRaw := "5s"

	fs.StringVar(&cfg.CIDRFile, "cidr-file", "", "CIDR CSV path")
	fs.StringVar(&cfg.PortFile, "port-file", "", "Port CSV path")
	fs.StringVar(&cfg.Output, "output", "scan_results.csv", "output csv")
	fs.DurationVar(&cfg.Timeout, "timeout", 100*time.Millisecond, "dial timeout")
	fs.DurationVar(&cfg.Delay, "delay", 10*time.Millisecond, "dispatch delay")
	fs.IntVar(&cfg.BucketRate, "bucket-rate", 100, "bucket rate")
	fs.IntVar(&cfg.BucketCapacity, "bucket-capacity", 100, "bucket capacity")
	fs.IntVar(&cfg.Workers, "workers", 10, "worker count")
	fs.StringVar(&cfg.PressureAPI, "pressure-api", "http://localhost:8080/api/pressure", "pressure api")
	fs.StringVar(&pressureIntervalRaw, "pressure-interval", "5s", "pressure poll interval (duration or seconds)")
	fs.BoolVar(&cfg.DisableAPI, "disable-api", false, "disable pressure api")
	fs.StringVar(&cfg.Resume, "resume", "", "resume state file")
	fs.StringVar(&cfg.LogLevel, "log-level", "info", "debug|info|error")
	fs.StringVar(&cfg.Format, "format", "human", "human|json")
	fs.StringVar(&cfg.CIDRIPCol, "cidr-ip-col", "ip", "cidr csv ip column name")
	fs.StringVar(&cfg.CIDRIPCidrCol, "cidr-ip-cidr-col", "ip_cidr", "cidr csv ip_cidr column name")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if cfg.CIDRFile == "" || cfg.PortFile == "" {
		return Config{}, errors.New("-cidr-file and -port-file are required")
	}
	if strings.TrimSpace(cfg.CIDRIPCol) == "" || strings.TrimSpace(cfg.CIDRIPCidrCol) == "" {
		return Config{}, errors.New("-cidr-ip-col and -cidr-ip-cidr-col must be non-empty")
	}
	if cfg.Format != "human" && cfg.Format != "json" {
		return Config{}, errors.New("-format must be human or json")
	}
	if seconds, err := strconv.Atoi(pressureIntervalRaw); err == nil {
		cfg.PressureInterval = time.Duration(seconds) * time.Second
	} else {
		interval, parseErr := time.ParseDuration(pressureIntervalRaw)
		if parseErr != nil {
			return Config{}, errors.New("-pressure-interval must be duration like 5s or integer seconds")
		}
		cfg.PressureInterval = interval
	}
	if cfg.PressureInterval <= 0 {
		return Config{}, errors.New("-pressure-interval must be > 0")
	}

	return cfg, nil
}
