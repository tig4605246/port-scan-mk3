package input

import "testing"

func mustLoadRows(t *testing.T, rows []CIDRRecord) []CIDRRecord {
	t.Helper()
	for i := range rows {
		if err := rows[i].Parse(); err != nil {
			t.Fatalf("parse row[%d] failed: %v", i, err)
		}
	}
	return rows
}

func TestValidateIPRows_DuplicateIPIPCidrPairFatal(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected duplicate pair error")
	}
}

func TestValidateIPRows_IPNotInsideIPCidrFatal(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.1.0.1", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected containment error")
	}
}

func TestValidateIPRows_SelectorCIDROutsideIPCidrFatal(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.1.0/24", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected selector containment error")
	}
}

func TestValidateIPRows_DifferentIPCidrOverlapFatal(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.1.1", IPCidrRaw: "10.0.0.0/16"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected ip_cidr overlap error")
	}
}

func TestValidateIPRows_SameIPCidrDifferentIPSelectorsOverlapFatal(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.0.1/32", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err == nil {
		t.Fatal("expected selector overlap error")
	}
}

func TestValidateIPRows_SameIPCidrWithDistinctSelectorsAllowed(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.0.2", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.0.8/30", IPCidrRaw: "10.0.0.0/24"},
	})
	if err := ValidateIPRows(rows); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestValidateNoOverlap_Wrapper(t *testing.T) {
	rows := mustLoadRows(t, []CIDRRecord{
		{IPRaw: "10.0.0.1", IPCidrRaw: "10.0.0.0/24"},
		{IPRaw: "10.0.1.1", IPCidrRaw: "10.0.1.0/24"},
	})
	if err := ValidateNoOverlap(rows); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
