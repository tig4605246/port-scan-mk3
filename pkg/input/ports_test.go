package input

import (
	"strings"
	"testing"
)

func TestLoadPorts_WhenRowsUseTCP_ParsesPorts(t *testing.T) {
	ports, err := LoadPorts(strings.NewReader("80/tcp\n443/tcp\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 2 || ports[0].Number != 80 {
		t.Fatalf("unexpected ports: %#v", ports)
	}
}
