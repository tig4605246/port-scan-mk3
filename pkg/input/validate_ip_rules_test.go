package input

import (
	"strings"
	"testing"
)

func mustLoadRows(t *testing.T, rows []CIDRRecord) []CIDRRecord {
	t.Helper()
	for i := range rows {
		if err := rows[i].Parse(); err != nil {
			t.Fatalf("parse row[%d] failed: %v", i, err)
		}
	}
	return rows
}

func TestValidateIPRows_WhenDuplicateIPAndIPCidrAndPortExists_ReturnsError(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24", Port: 443},
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24", Port: 443},
	})
	err := ValidateIPRows(rows)
	if err == nil {
		t.Fatal("expected duplicate tuple error")
	}
	if !strings.Contains(err.Error(), "duplicate (ip,ip_cidr,port)") {
		t.Fatalf("expected duplicate tuple error, got %v", err)
	}
}

func TestDuplicateRowKey_WhenOnlyPortDiffers_ReturnsDifferentKeys(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24", Port: 80},
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24", Port: 443},
	})

	if duplicateRowKey(rows[0]) == duplicateRowKey(rows[1]) {
		t.Fatal("expected duplicate key to include port")
	}
}

func TestValidateIPRows_WhenIPIsOutsideIPCidr_ReturnsError(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.1.0.1", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected containment error")
	}
}

func TestValidateIPRows_WhenSelectorCIDRIsOutsideIPCidr_ReturnsError(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.1.0/24", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected selector containment error")
	}
}

func TestValidateIPRows_WhenDifferentIPCidrsOverlap_ReturnsError(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.1.1", IPCidrRaw: "10.0.0.0/16"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected ip_cidr overlap error")
	}
}

func TestValidateIPRows_WhenSelectorsOverlapWithinSameIPCidr_ReturnsError(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.0.1/32", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected selector overlap error")
	}
}

func TestValidateIPRows_WhenSelectorsAreDistinctWithinSameIPCidr_ReturnsNil(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.0.2", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.0.8/30", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestValidateNoOverlap_WhenNetworksDoNotOverlap_ReturnsNil(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.1.1", IPCidrRaw: "10.0.1.0/24"},
	})
	if err := ValidateNoOverlap(rows); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
