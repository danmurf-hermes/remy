package memory

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func now() int64 {
	return time.Now().UnixMilli()
}

func writeJSON(w io.Writer, v any) error {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encoding json: %w", err)
	}
	return nil
}

func readJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf("decoding json: %w", err)
	}
	return nil
}
