// Package hook implements helpers for the cc2md hook subcommand,
// including parsing the Stop-hook stdin payload.
package hook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Input is the JSON payload Claude Code sends to a Stop hook on stdin.
// Unknown fields in the payload are silently ignored.
type Input struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
}

// ParseInput decodes a Claude Code Stop-hook JSON payload from r and
// returns the resulting Input. It returns an error if the JSON cannot be
// decoded or if transcript_path is missing or empty.
func ParseInput(r io.Reader) (Input, error) {
	var in Input
	dec := json.NewDecoder(r)
	if err := dec.Decode(&in); err != nil {
		return Input{}, fmt.Errorf("decode hook input: %w", err)
	}
	if in.TranscriptPath == "" {
		return Input{}, errors.New("hook input: missing transcript_path")
	}
	return in, nil
}
