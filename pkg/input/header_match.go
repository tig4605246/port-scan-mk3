package input

import "strings"

func canonicalHeaderName(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func detectRichHeaderIndices(header []string) (map[string]int, bool) {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[canonicalHeaderName(h)] = i
	}
	for _, field := range requiredRichFields {
		if _, ok := idx[field]; !ok {
			return nil, false
		}
	}
	return idx, true
}
