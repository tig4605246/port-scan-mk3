package input

import (
	"strings"
	"testing"
)

func TestLoadCIDRs_WhenHeaderIsInvalid_ReturnsError(t *testing.T) {
	_, err := LoadCIDRs(strings.NewReader("a,b,c\n1,10.0.0.1,10.0.0.0/24\n"))
	if err == nil {
		t.Fatal("expected header error")
	}
}

func TestLoadPorts_WhenProtocolIsNotTCP_ReturnsError(t *testing.T) {
	_, err := LoadPorts(strings.NewReader("53/udp\n"))
	if err == nil {
		t.Fatal("expected invalid protocol error")
	}
}
