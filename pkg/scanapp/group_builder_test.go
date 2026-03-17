package scanapp

import (
	"net"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/input"
)

func TestBuildGroups_WhenBasicStrategy_ProducesSameResultAsBuildCIDRGroups(t *testing.T) {
	_, ipNet, _ := net.ParseCIDR("10.0.0.0/30")
	records := []input.CIDRRecord{
		{FabName: "fab1", IPRaw: "10.0.0.1", CIDR: "10.0.0.0/30", CIDRName: "net-a", Net: ipNet},
		{FabName: "fab2", IPRaw: "10.0.0.2", CIDR: "10.0.0.0/30", CIDRName: "net-a", Net: ipNet},
	}

	groups, err := buildGroups(records, basicGroupStrategy{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	g := groups["10.0.0.0/30"]
	if len(g.targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(g.targets))
	}
}

func TestBuildGroups_WhenRichStrategy_ProducesSameResultAsBuildRichGroups(t *testing.T) {
	records := []input.CIDRRecord{
		{
			IsRich: true, IsValid: true, ExecutionKey: "10.0.0.1:80/tcp",
			DstIP: "10.0.0.1", DstNetworkSegment: "10.0.0.0/24", Port: 80,
			FabName: "fab1", CIDRName: "net-a",
		},
	}

	groups, err := buildGroups(records, richGroupStrategy{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	g := groups["10.0.0.1:80/tcp"]
	if len(g.targets) != 1 || g.targets[0].ip != "10.0.0.1" {
		t.Fatalf("unexpected target: %+v", g.targets)
	}
	if g.port != 80 {
		t.Fatalf("expected port 80, got %d", g.port)
	}
}
