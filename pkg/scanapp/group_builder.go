package scanapp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xuxiping/port-scan-mk3/pkg/input"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

type cidrGroup struct {
	targets []scanTarget
	port    int
}

type groupBuildStrategy interface {
	ShouldInclude(rec input.CIDRRecord) bool
	Key(rec input.CIDRRecord) (string, error)
	NewGroup(rec input.CIDRRecord) (cidrGroup, error)
	MergeGroup(existing cidrGroup, rec input.CIDRRecord) (cidrGroup, error)
	RequireNonEmpty() bool
}

func buildGroups(records []input.CIDRRecord, strategy groupBuildStrategy) (map[string]cidrGroup, error) {
	out := make(map[string]cidrGroup)
	for _, rec := range records {
		if !strategy.ShouldInclude(rec) {
			continue
		}

		key, err := strategy.Key(rec)
		if err != nil {
			return nil, err
		}

		group, ok := out[key]
		if !ok {
			group, err = strategy.NewGroup(rec)
			if err != nil {
				return nil, err
			}
		} else {
			group, err = strategy.MergeGroup(group, rec)
			if err != nil {
				return nil, err
			}
		}
		out[key] = group
	}

	if len(out) == 0 && strategy.RequireNonEmpty() {
		return nil, fmt.Errorf("no usable input rows")
	}

	for key, group := range out {
		sort.Slice(group.targets, func(i, j int) bool {
			return ipv4ToUint32(group.targets[i].ip) < ipv4ToUint32(group.targets[j].ip)
		})
		out[key] = group
	}

	return out, nil
}

type basicGroupStrategy struct{}

func (basicGroupStrategy) ShouldInclude(_ input.CIDRRecord) bool { return true }

func (basicGroupStrategy) Key(rec input.CIDRRecord) (string, error) {
	cidr := rec.CIDR
	if cidr == "" && rec.Net != nil {
		cidr = rec.Net.String()
	}
	if cidr == "" {
		return "", fmt.Errorf("record missing ip_cidr")
	}
	return cidr, nil
}

func (s basicGroupStrategy) NewGroup(rec input.CIDRRecord) (cidrGroup, error) {
	targets, err := s.targets(rec)
	if err != nil {
		return cidrGroup{}, err
	}
	return cidrGroup{targets: targets}, nil
}

func (s basicGroupStrategy) MergeGroup(existing cidrGroup, rec input.CIDRRecord) (cidrGroup, error) {
	targets, err := s.targets(rec)
	if err != nil {
		return cidrGroup{}, err
	}
	existing.targets = append(existing.targets, targets...)
	return existing, nil
}

func (basicGroupStrategy) RequireNonEmpty() bool { return false }

func (basicGroupStrategy) targets(rec input.CIDRRecord) ([]scanTarget, error) {
	cidr := rec.CIDR
	if cidr == "" && rec.Net != nil {
		cidr = rec.Net.String()
	}

	selector := ""
	switch {
	case rec.Selector != nil:
		selector = rec.Selector.String()
	case strings.TrimSpace(rec.IPRaw) != "":
		selector = strings.TrimSpace(rec.IPRaw)
	case rec.Net != nil:
		selector = rec.Net.String()
	default:
		return nil, fmt.Errorf("record for cidr %s missing selector", cidr)
	}

	ips, err := task.ExpandIPSelectors([]string{selector})
	if err != nil {
		return nil, fmt.Errorf("expand selector failed for cidr %s: %w", cidr, err)
	}

	targets := make([]scanTarget, 0, len(ips))
	for _, ip := range ips {
		targets = append(targets, scanTarget{
			ip: ip,
			meta: targetMeta{
				fabName:  rec.FabName,
				cidrName: rec.CIDRName,
			},
		})
	}
	return targets, nil
}

type richGroupStrategy struct{}

func (richGroupStrategy) ShouldInclude(rec input.CIDRRecord) bool {
	return rec.IsRich && rec.IsValid
}

func (richGroupStrategy) Key(rec input.CIDRRecord) (string, error) {
	key := strings.TrimSpace(rec.ExecutionKey)
	if key == "" {
		return "", fmt.Errorf("rich record missing execution_key at row %d", rec.RowNumber)
	}
	return key, nil
}

func (richGroupStrategy) NewGroup(rec input.CIDRRecord) (cidrGroup, error) {
	key, err := (richGroupStrategy{}).Key(rec)
	if err != nil {
		return cidrGroup{}, err
	}

	return cidrGroup{
		port: rec.Port,
		targets: []scanTarget{{
			ip:     rec.DstIP,
			ipCidr: rec.DstNetworkSegment,
			meta: targetMeta{
				fabName:           rec.FabName,
				cidrName:          rec.CIDRName,
				serviceLabel:      rec.ServiceLabel,
				decision:          rec.Decision,
				policyID:          rec.PolicyID,
				reason:            rec.Reason,
				executionKey:      key,
				srcIP:             rec.SrcIP,
				srcNetworkSegment: rec.SrcNetworkSegment,
			},
		}},
	}, nil
}

func (richGroupStrategy) MergeGroup(existing cidrGroup, rec input.CIDRRecord) (cidrGroup, error) {
	key := strings.TrimSpace(rec.ExecutionKey)
	if existing.port != rec.Port {
		return cidrGroup{}, fmt.Errorf("execution key %s has inconsistent port", key)
	}

	target := &existing.targets[0]
	target.meta.fabName = mergeFieldValue(target.meta.fabName, rec.FabName)
	target.meta.cidrName = mergeFieldValue(target.meta.cidrName, rec.CIDRName)
	target.meta.serviceLabel = mergeFieldValue(target.meta.serviceLabel, rec.ServiceLabel)
	target.meta.decision = mergeFieldValue(target.meta.decision, rec.Decision)
	target.meta.policyID = mergeFieldValue(target.meta.policyID, rec.PolicyID)
	target.meta.reason = mergeFieldValue(target.meta.reason, rec.Reason)
	target.meta.srcIP = mergeFieldValue(target.meta.srcIP, rec.SrcIP)
	target.meta.srcNetworkSegment = mergeFieldValue(target.meta.srcNetworkSegment, rec.SrcNetworkSegment)
	return existing, nil
}

func (richGroupStrategy) RequireNonEmpty() bool { return true }

func mergeFieldValue(existing, incoming string) string {
	existing = strings.TrimSpace(existing)
	incoming = strings.TrimSpace(incoming)
	if incoming == "" || existing == incoming {
		return existing
	}
	if existing == "" {
		return incoming
	}
	parts := strings.Split(existing, "|")
	for _, p := range parts {
		if p == incoming {
			return existing
		}
	}
	return existing + "|" + incoming
}

func buildCIDRGroups(cidrRecords []input.CIDRRecord) (map[string]cidrGroup, error) {
	return buildGroups(cidrRecords, basicGroupStrategy{})
}

func buildRichGroups(cidrRecords []input.CIDRRecord) (map[string]cidrGroup, error) {
	return buildGroups(cidrRecords, richGroupStrategy{})
}
