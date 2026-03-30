package scanapp

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ReachabilityResult struct {
	IP          string
	Reachable   bool
	FailureText string
}

type ReachabilityChecker interface {
	Check(ctx context.Context, ip string, timeout time.Duration) ReachabilityResult
}

type commandReachabilityChecker struct {
	goos   string
	runner commandRunner
}

type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, name string, args ...string) error {
	return exec.CommandContext(ctx, name, args...).Run()
}

func (c *commandReachabilityChecker) Check(ctx context.Context, ip string, timeout time.Duration) ReachabilityResult {
	result := ReachabilityResult{IP: ip}
	if strings.TrimSpace(ip) == "" {
		result.FailureText = "ip is required"
		return result
	}

	name, args, err := buildPingCommand(c.goos, ip, timeout)
	if err != nil {
		result.FailureText = err.Error()
		return result
	}

	runner := c.runner
	if runner == nil {
		runner = execCommandRunner{}
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := runner.Run(runCtx, name, args...); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			result.FailureText = err.Error()
			return result
		}
		result.FailureText = err.Error()
		return result
	}

	result.Reachable = true
	return result
}

func buildPingCommand(goos, ip string, timeout time.Duration) (string, []string, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "", nil, errors.New("ip is required")
	}

	if goos == "" {
		goos = runtime.GOOS
	}

	if goos == "windows" {
		timeoutMS := timeout.Milliseconds()
		if timeoutMS < 0 {
			timeoutMS = 0
		}
		return "ping", []string{"-n", "1", "-w", strconv.FormatInt(timeoutMS, 10), ip}, nil
	}

	return "ping", []string{"-c", "1", ip}, nil
}
