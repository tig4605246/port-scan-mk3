package input

import "net"

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

type PortSpec struct {
	Number int
	Proto  string
	Raw    string
}
