package input

const (
	RichFieldSrcIP             = "src_ip"
	RichFieldSrcNetworkSegment = "src_network_segment"
	RichFieldDstIP             = "dst_ip"
	RichFieldDstNetworkSegment = "dst_network_segment"
	RichFieldServiceLabel      = "service_label"
	RichFieldProtocol          = "protocol"
	RichFieldPort              = "port"
	RichFieldDecision          = "decision"
	RichFieldPolicyID          = "policy_id"
	RichFieldReason            = "reason"
)

var requiredRichFields = []string{
	RichFieldSrcIP,
	RichFieldSrcNetworkSegment,
	RichFieldDstIP,
	RichFieldDstNetworkSegment,
	RichFieldServiceLabel,
	RichFieldProtocol,
	RichFieldPort,
	RichFieldDecision,
	RichFieldPolicyID,
	RichFieldReason,
}

// RichParseSummary keeps row-level parse outcomes for rich input mode.
type RichParseSummary struct {
	TotalRows       int
	ValidRows       int
	InvalidRows     int
	FailureByReason map[string]int
}
