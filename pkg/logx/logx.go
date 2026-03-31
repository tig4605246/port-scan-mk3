package logx

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

func LogJSON(out io.Writer, level, msg string, fields map[string]any) {
	payload := map[string]any{
		"level":  level,
		"msg":    msg,
		"fields": fields,
		"ts":     time.Now().UTC().Format(time.RFC3339),
	}
	if err := json.NewEncoder(out).Encode(payload); err != nil {
		fmt.Fprintf(os.Stderr, "logx: json encode failed: %v\n", err)
	}
}
