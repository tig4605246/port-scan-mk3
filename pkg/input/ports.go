package input

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func LoadPorts(r io.Reader) ([]PortSpec, error) {
	scanner := bufio.NewScanner(r)
	out := make([]PortSpec, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, "/")
		if len(parts) != 2 || strings.ToLower(parts[1]) != "tcp" {
			return nil, fmt.Errorf("invalid port row: %s", line)
		}
		n, err := strconv.Atoi(parts[0])
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("invalid port number: %s", line)
		}
		out = append(out, PortSpec{Number: n, Proto: "tcp", Raw: line})
	}
	return out, scanner.Err()
}
