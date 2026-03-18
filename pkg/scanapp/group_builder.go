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
	return richCIDRKey(rec)
}

func (richGroupStrategy) NewGroup(rec input.CIDRRecord) (cidrGroup, error) {
	target, err := richTargetFromRecord(rec)
	if err != nil {
		return cidrGroup{}, err
	}
	return cidrGroup{
		targets: []scanTarget{target},
	}, nil
}

func (richGroupStrategy) MergeGroup(existing cidrGroup, rec input.CIDRRecord) (cidrGroup, error) {
	key := strings.TrimSpace(rec.ExecutionKey)
	if key == "" {
		return cidrGroup{}, fmt.Errorf("rich record missing execution_key at row %d", rec.RowNumber)
	}
	for i := range existing.targets {
		target := &existing.targets[i]
		if strings.TrimSpace(target.meta.executionKey) != key {
			continue
		}
		if target.port != rec.Port {
			return cidrGroup{}, fmt.Errorf("execution key %s has inconsistent port", key)
		}
		mergeRichMetadataFromRecord(target, rec)
		return existing, nil
	}
	target, err := richTargetFromRecord(rec)
	if err != nil {
		return cidrGroup{}, err
	}
	existing.targets = append(existing.targets, target)
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
	groups, err := buildGroups(cidrRecords, richGroupStrategy{})
	if err != nil {
		return nil, err
	}

	ownerByExecutionKey := make(map[string]string)
	for _, rec := range cidrRecords {
		if !rec.IsRich || !rec.IsValid {
			continue
		}
		key := strings.TrimSpace(rec.ExecutionKey)
		if key == "" {
			return nil, fmt.Errorf("rich record missing execution_key at row %d", rec.RowNumber)
		}
		cidr, err := richCIDRKey(rec)
		if err != nil {
			return nil, err
		}
		if _, ok := ownerByExecutionKey[key]; !ok {
			ownerByExecutionKey[key] = cidr
		}
	}

	for cidr, group := range groups {
		kept := make([]scanTarget, 0, len(group.targets))
		for _, target := range group.targets {
			executionKey := strings.TrimSpace(target.meta.executionKey)
			ownerCIDR, ok := ownerByExecutionKey[executionKey]
			if !ok {
				return nil, fmt.Errorf("owner cidr missing for execution key %s", executionKey)
			}
			if ownerCIDR == cidr {
				kept = append(kept, target)
				continue
			}

			ownerGroup, ok := groups[ownerCIDR]
			if !ok {
				return nil, fmt.Errorf("owner cidr %s missing for execution key %s", ownerCIDR, executionKey)
			}
			idx := richTargetIndexByExecutionKey(ownerGroup.targets, executionKey)
			if idx < 0 {
				ownerGroup.targets = append(ownerGroup.targets, target)
			} else if err := mergeRichTargetValues(&ownerGroup.targets[idx], target); err != nil {
				return nil, err
			}
			groups[ownerCIDR] = ownerGroup
		}
		if len(kept) == 0 {
			delete(groups, cidr)
			continue
		}
		group.targets = kept
		groups[cidr] = group
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("no usable input rows")
	}

	for cidr, group := range groups {
		sort.Slice(group.targets, func(i, j int) bool {
			left := group.targets[i]
			right := group.targets[j]
			leftIP := ipv4ToUint32(left.ip)
			rightIP := ipv4ToUint32(right.ip)
			if leftIP != rightIP {
				return leftIP < rightIP
			}
			if left.port != right.port {
				return left.port < right.port
			}
			return strings.TrimSpace(left.meta.executionKey) < strings.TrimSpace(right.meta.executionKey)
		})
		groups[cidr] = group
	}
	return groups, nil
}

func richCIDRKey(rec input.CIDRRecord) (string, error) {
	if cidr := strings.TrimSpace(rec.DstNetworkSegment); cidr != "" {
		return cidr, nil
	}
	if cidr := strings.TrimSpace(rec.CIDR); cidr != "" {
		return cidr, nil
	}
	if rec.Net != nil {
		return rec.Net.String(), nil
	}
	return "", fmt.Errorf("rich record missing dst_network_segment at row %d", rec.RowNumber)
}

func richTargetFromRecord(rec input.CIDRRecord) (scanTarget, error) {
	key := strings.TrimSpace(rec.ExecutionKey)
	if key == "" {
		return scanTarget{}, fmt.Errorf("rich record missing execution_key at row %d", rec.RowNumber)
	}
	cidr, err := richCIDRKey(rec)
	if err != nil {
		return scanTarget{}, err
	}
	return scanTarget{
		ip:     rec.DstIP,
		ipCidr: cidr,
		port:   rec.Port,
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
	}, nil
}

func mergeRichMetadataFromRecord(target *scanTarget, rec input.CIDRRecord) {
	target.meta.fabName = mergeFieldValue(target.meta.fabName, rec.FabName)
	target.meta.cidrName = mergeFieldValue(target.meta.cidrName, rec.CIDRName)
	target.meta.serviceLabel = mergeFieldValue(target.meta.serviceLabel, rec.ServiceLabel)
	target.meta.decision = mergeFieldValue(target.meta.decision, rec.Decision)
	target.meta.policyID = mergeFieldValue(target.meta.policyID, rec.PolicyID)
	target.meta.reason = mergeFieldValue(target.meta.reason, rec.Reason)
	target.meta.srcIP = mergeFieldValue(target.meta.srcIP, rec.SrcIP)
	target.meta.srcNetworkSegment = mergeFieldValue(target.meta.srcNetworkSegment, rec.SrcNetworkSegment)
}

func richTargetIndexByExecutionKey(targets []scanTarget, executionKey string) int {
	for i := range targets {
		if strings.TrimSpace(targets[i].meta.executionKey) == executionKey {
			return i
		}
	}
	return -1
}

func mergeRichTargetValues(dst *scanTarget, incoming scanTarget) error {
	key := strings.TrimSpace(dst.meta.executionKey)
	if key == "" {
		return fmt.Errorf("destination rich target missing execution key")
	}
	if strings.TrimSpace(incoming.meta.executionKey) != key {
		return fmt.Errorf("cannot merge rich targets with different execution keys: %s vs %s", key, strings.TrimSpace(incoming.meta.executionKey))
	}
	if dst.port != incoming.port {
		return fmt.Errorf("execution key %s has inconsistent port", key)
	}
	mergeRichMetadataFromRecord(dst, input.CIDRRecord{
		FabName:           incoming.meta.fabName,
		CIDRName:          incoming.meta.cidrName,
		ServiceLabel:      incoming.meta.serviceLabel,
		Decision:          incoming.meta.decision,
		PolicyID:          incoming.meta.policyID,
		Reason:            incoming.meta.reason,
		SrcIP:             incoming.meta.srcIP,
		SrcNetworkSegment: incoming.meta.srcNetworkSegment,
	})
	return nil
}
