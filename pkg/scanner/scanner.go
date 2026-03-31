package scanner

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"
)

type Result struct {
	IP             string
	Port           int
	Status         string
	ResponseTimeMS int64
	Error          string
}

func ScanTCP(dial func(context.Context, string, string) (net.Conn, error), ip string, port int, timeout time.Duration) Result {
	target := net.JoinHostPort(ip, strconv.Itoa(port))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	conn, err := dial(ctx, "tcp", target)
	if err == nil {
		if err := conn.Close(); err != nil {
			// non-fatal: connection already closed; log if needed
		}
		return Result{
			IP:             ip,
			Port:           port,
			Status:         "open",
			ResponseTimeMS: time.Since(start).Milliseconds(),
		}
	}

	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return Result{
			IP:             ip,
			Port:           port,
			Status:         "close(timeout)",
			ResponseTimeMS: 0,
			Error:          err.Error(),
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return Result{
			IP:             ip,
			Port:           port,
			Status:         "close(timeout)",
			ResponseTimeMS: 0,
			Error:          err.Error(),
		}
	}

	return Result{
		IP:             ip,
		Port:           port,
		Status:         "close",
		ResponseTimeMS: 0,
		Error:          err.Error(),
	}
}
