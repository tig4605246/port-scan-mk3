package input

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

// ParseRichRows parses and validates rich CIDR input rows. It keeps one
// row-level result per source row (valid or invalid). Valid rows are converted
// into CIDRRecord values consumable by downstream scan runtime code.
func ParseRichRows(rows [][]string, idx map[string]int) ([]CIDRRecord, RichParseSummary, error) {
	summary := RichParseSummary{
		TotalRows:       max(0, len(rows)-1),
		FailureByReason: map[string]int{},
	}
	out := make([]CIDRRecord, 0, summary.TotalRows)
	for i := 1; i < len(rows); i++ {
		rec, code, err := parseRichRow(rows[i], i+1, idx)
		if err != nil {
			summary.InvalidRows++
			summary.FailureByReason[code]++
			out = append(out, CIDRRecord{
				RowNumber:       i + 1,
				IsRich:          true,
				IsValid:         false,
				ValidationCode:  code,
				ValidationError: err.Error(),
			})
			continue
		}
		summary.ValidRows++
		out = append(out, rec)
	}
	if summary.ValidRows == 0 {
		return out, summary, fmt.Errorf("no usable input rows")
	}
	return out, summary, nil
}

func parseRichRow(row []string, rowNumber int, idx map[string]int) (CIDRRecord, string, error) {
	get := func(field string) string {
		i := idx[field]
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	srcIPRaw := get(RichFieldSrcIP)
	srcSegRaw := get(RichFieldSrcNetworkSegment)
	dstIPRaw := get(RichFieldDstIP)
	dstSegRaw := get(RichFieldDstNetworkSegment)
	serviceLabel := get(RichFieldServiceLabel)
	protocolRaw := get(RichFieldProtocol)
	portRaw := get(RichFieldPort)
	decisionRaw := get(RichFieldDecision)
	policyID := get(RichFieldPolicyID)
	reason := get(RichFieldReason)

	required := map[string]string{
		RichFieldSrcIP:             srcIPRaw,
		RichFieldSrcNetworkSegment: srcSegRaw,
		RichFieldDstIP:             dstIPRaw,
		RichFieldDstNetworkSegment: dstSegRaw,
		RichFieldServiceLabel:      serviceLabel,
		RichFieldProtocol:          protocolRaw,
		RichFieldPort:              portRaw,
		RichFieldDecision:          decisionRaw,
		RichFieldPolicyID:          policyID,
		RichFieldReason:            reason,
	}
	for field, value := range required {
		if value == "" {
			return CIDRRecord{}, ValidationMissingField, fmt.Errorf("missing required field %s", field)
		}
	}

	srcIP := net.ParseIP(srcIPRaw).To4()
	if srcIP == nil {
		return CIDRRecord{}, ValidationInvalidSrcIP, fmt.Errorf("invalid src_ip %q", srcIPRaw)
	}
	dstIP := net.ParseIP(dstIPRaw).To4()
	if dstIP == nil {
		return CIDRRecord{}, ValidationInvalidDstIP, fmt.Errorf("invalid dst_ip %q", dstIPRaw)
	}

	_, srcSeg, err := net.ParseCIDR(srcSegRaw)
	if err != nil || srcSeg.IP.To4() == nil {
		return CIDRRecord{}, ValidationInvalidSrcSegment, fmt.Errorf("invalid src_network_segment %q", srcSegRaw)
	}
	_, dstSeg, err := net.ParseCIDR(dstSegRaw)
	if err != nil || dstSeg.IP.To4() == nil {
		return CIDRRecord{}, ValidationInvalidDstSegment, fmt.Errorf("invalid dst_network_segment %q", dstSegRaw)
	}
	if !srcSeg.Contains(srcIP) {
		return CIDRRecord{}, ValidationSrcContainmentFail, fmt.Errorf("src_ip %s not in src_network_segment %s", srcIP.String(), srcSeg.String())
	}
	if !dstSeg.Contains(dstIP) {
		return CIDRRecord{}, ValidationDstContainmentFail, fmt.Errorf("dst_ip %s not in dst_network_segment %s", dstIP.String(), dstSeg.String())
	}

	protocol := strings.ToLower(protocolRaw)
	if protocol != "tcp" {
		return CIDRRecord{}, ValidationInvalidProtocol, fmt.Errorf("invalid protocol %q", protocolRaw)
	}
	decision := strings.ToLower(decisionRaw)
	if decision != "accept" && decision != "deny" {
		return CIDRRecord{}, ValidationInvalidDecision, fmt.Errorf("invalid decision %q", decisionRaw)
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil || port < 1 || port > 65535 {
		return CIDRRecord{}, ValidationInvalidPort, fmt.Errorf("invalid port %q", portRaw)
	}
	key, err := task.BuildExecutionKey(dstIP.String(), port, protocol)
	if err != nil {
		return CIDRRecord{}, ValidationInvalidPort, err
	}

	rec := CIDRRecord{
		FabName:             srcIP.String(),
		CIDRName:            serviceLabel,
		IPRaw:               dstIP.String(),
		IPCidrRaw:           dstSeg.String(),
		RowNumber:           rowNumber,
		IPColName:           RichFieldDstIP,
		IPCidrCol:           RichFieldDstNetworkSegment,
		IsRich:              true,
		IsValid:             true,
		SrcIP:               srcIP.String(),
		SrcNetworkSegment:   srcSeg.String(),
		DstIP:               dstIP.String(),
		DstNetworkSegment:   dstSeg.String(),
		ServiceLabel:        serviceLabel,
		Protocol:            protocol,
		Port:                port,
		Decision:            decision,
		PolicyID:            policyID,
		Reason:              reason,
		ExecutionKey:        key,
		RichInputIdentifier: fmt.Sprintf("row:%d", rowNumber),
	}
	if err := rec.Parse(); err != nil {
		return CIDRRecord{}, ValidationInvalidDstIP, err
	}
	return rec, "", nil
}
