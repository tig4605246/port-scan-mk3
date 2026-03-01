package input

import "net"

type CIDRRecord struct {
	FabName  string
	CIDR     string
	CIDRName string
	Net      *net.IPNet
}

type PortSpec struct {
	Number int
	Proto  string
	Raw    string
}
