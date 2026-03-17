package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
)

func TestInputs_WhenFilesAreValid_ReturnsValidResult(t *testing.T) {
	tmp := t.TempDir()
	cidr := filepath.Join(tmp, "cidr.csv")
	port := filepath.Join(tmp, "ports.csv")
	if err := os.WriteFile(cidr, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,10.0.0.1,10.0.0.0/24,a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(port, []byte("80/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := Inputs(config.Config{
		CIDRFile:      cidr,
		PortFile:      port,
		CIDRIPCol:     "ip",
		CIDRIPCidrCol: "ip_cidr",
	})

	if !result.Valid || result.Detail != "ok" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestInputs_WhenCIDRRowsInvalid_ReturnsInvalidDetail(t *testing.T) {
	tmp := t.TempDir()
	cidr := filepath.Join(tmp, "cidr.csv")
	port := filepath.Join(tmp, "ports.csv")
	if err := os.WriteFile(cidr, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,10.0.1.1,10.0.0.0/24,a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(port, []byte("80/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := Inputs(config.Config{
		CIDRFile:      cidr,
		PortFile:      port,
		CIDRIPCol:     "ip",
		CIDRIPCidrCol: "ip_cidr",
	})

	if result.Valid {
		t.Fatalf("expected invalid result, got %+v", result)
	}
	if !strings.Contains(result.Detail, "outside ip_cidr") {
		t.Fatalf("unexpected detail: %s", result.Detail)
	}
}
