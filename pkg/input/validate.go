package input

import "fmt"

func ValidateNoOverlap(networks []CIDRRecord) error {
	for i := 0; i < len(networks); i++ {
		for j := i + 1; j < len(networks); j++ {
			a := networks[i].Net
			b := networks[j].Net
			if a.Contains(b.IP) || b.Contains(a.IP) {
				return fmt.Errorf("overlap detected: %s (%s) <-> %s (%s)",
					networks[i].CIDRName, networks[i].CIDR,
					networks[j].CIDRName, networks[j].CIDR)
			}
		}
	}
	return nil
}
