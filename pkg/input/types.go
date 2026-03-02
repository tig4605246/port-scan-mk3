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
}

// PortSpec is one normalized TCP port row from the port input file.
type PortSpec struct {
	Number int
	Proto  string
	Raw    string
}
