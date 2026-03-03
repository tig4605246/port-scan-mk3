package input

import "net"

// CIDRRecord is one parsed row from the CIDR CSV input.
// It keeps the original selector/cidr strings and normalized network forms.
type CIDRRecord struct {
	FabName   string
	CIDR      string
	CIDRName  string
	Net       *net.IPNet
	IPRaw     string
	IPCidrRaw string
	Selector  *net.IPNet
	RowNumber int
	IPColName string
	IPCidrCol string

	// Rich input fields.
	IsRich              bool
	IsValid             bool
	ValidationCode      string
	ValidationError     string
	SrcIP               string
	SrcNetworkSegment   string
	DstIP               string
	DstNetworkSegment   string
	ServiceLabel        string
	Protocol            string
	Port                int
	Decision            string
	PolicyID            string
	Reason              string
	ExecutionKey        string
	RichInputIdentifier string
}

// PortSpec is one normalized TCP port row from the port input file.
type PortSpec struct {
	Number int
	Proto  string
	Raw    string
}
