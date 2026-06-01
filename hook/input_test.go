package hook

import (
	"strings"
	"testing"
)

func TestParseInput_Valid(t *testing.T) {
	const payload = `{
		"session_id": "0b9c1f3a-7e4d-4f2a-b8c1-3d4e5f6a7b8c",
		"transcript_path": "/tmp/example.jsonl",
		"cwd": "/Users/ryan/Projects/cc2md",
		"hook_event_name": "Stop"
	}`

	in, err := ParseInput(strings.NewReader(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.SessionID != "0b9c1f3a-7e4d-4f2a-b8c1-3d4e5f6a7b8c" {
		t.Errorf("SessionID: got %q", in.SessionID)
	}
	if in.TranscriptPath != "/tmp/example.jsonl" {
		t.Errorf("TranscriptPath: got %q", in.TranscriptPath)
	}
	if in.CWD != "/Users/ryan/Projects/cc2md" {
		t.Errorf("CWD: got %q", in.CWD)
	}
	if in.HookEventName != "Stop" {
		t.Errorf("HookEventName: got %q", in.HookEventName)
	}
}

func TestParseInput_MissingTranscriptPath(t *testing.T) {
	const payload = `{"session_id": "x", "hook_event_name": "Stop"}`
	_, err := ParseInput(strings.NewReader(payload))
	if err == nil {
		t.Fatal("expected error for missing transcript_path, got nil")
	}
	if !strings.Contains(err.Error(), "transcript_path") {
		t.Errorf("expected error to mention transcript_path; got %v", err)
	}
}

func TestParseInput_UnknownFieldsIgnored(t *testing.T) {
	const payload = `{
		"session_id": "x",
		"transcript_path": "/tmp/x.jsonl",
		"some_future_field": 42
	}`
	in, err := ParseInput(strings.NewReader(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.SessionID != "x" || in.TranscriptPath != "/tmp/x.jsonl" {
		t.Errorf("unexpected struct: %+v", in)
	}
}

func TestParseInput_MalformedJSON(t *testing.T) {
	_, err := ParseInput(strings.NewReader("{not json"))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}
