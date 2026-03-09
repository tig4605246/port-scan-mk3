package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteValidation preserves the public validate output contract for human and JSON callers.
func WriteValidation(out io.Writer, format string, valid bool, detail string) error {
	if format == "json" {
		return json.NewEncoder(out).Encode(map[string]any{
			"valid":  valid,
			"detail": detail,
		})
	}
	_, err := fmt.Fprintf(out, "valid=%t detail=%s\n", valid, detail)
	return err
}
