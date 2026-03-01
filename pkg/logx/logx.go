package logx

import (
	"encoding/json"
	"io"
	"time"
)

func LogJSON(out io.Writer, level, msg string, fields map[string]any) {
	payload := map[string]any{
		"level":  level,
		"msg":    msg,
		"fields": fields,
		"ts":     time.Now().UTC().Format(time.RFC3339),
	}
	_ = json.NewEncoder(out).Encode(payload)
}
