package main

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/event"
)

// TestTheStatuslineCarriesTheWindow: the context event made from Claude
// Code's statusline payload carries the model's window along with the
// reading, and NOT the payload's token "totals", which are the window's
// tokens from the last response and not the session's spend; a payload with
// no reading makes no event.
func TestTheStatuslineCarriesTheWindow(t *testing.T) {
	in := []byte(`{"session_id":"s1","context_window":{"total_input_tokens":1200000,"total_output_tokens":34567,"context_window_size":1000000,"used_percentage":38.5}}`)
	e, ok := statuslineEvent(in)
	if !ok || e.Kind != event.Context || e.Frac == nil || *e.Frac != 0.385 {
		t.Fatalf("event %+v ok=%v", e, ok)
	}
	if e.Tokens != 0 || e.Window != 1_000_000 || e.Session != "s1" {
		t.Errorf("tokens %d window %d session %q", e.Tokens, e.Window, e.Session)
	}
	if _, ok := statuslineEvent([]byte(`{"context_window":{"used_percentage":null}}`)); ok {
		t.Error("a payload with no reading made an event")
	}
}
