package scanner

import (
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

func ScanTCP(dial func(string, string, time.Duration) (net.Conn, error), ip string, port int, timeout time.Duration) Result {
	target := net.JoinHostPort(ip, strconv.Itoa(port))
	start := time.Now()
	conn, err := dial("tcp", target, timeout)
	if err == nil {
		_ = conn.Close()
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

	return Result{
		IP:             ip,
		Port:           port,
		Status:         "close",
		ResponseTimeMS: 0,
		Error:          err.Error(),
	}
}
