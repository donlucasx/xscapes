package term

import "testing"

// Every terminal gets the cube. The scene is designed for it, and a truecolor
// terminal painting the palette's raw blends showed a different picture (a tan
// sun where Terminal.app shows peach over cream, notes/sunprobe, 2026-09-04).
// COLORTERM was already untrustworthy (Terminal.app exports truecolor inside
// Claude Code); now it is not consulted at all. XSCAPES_COLOR still opts in.
func TestEveryTerminalGetsTheCube(t *testing.T) {
	t.Setenv("XSCAPES_COLOR", "")
	t.Setenv("ASCIISCAPES_COLOR", "")
	for _, tc := range []struct{ prog, colorterm string }{
		{"Apple_Terminal", "truecolor"},
		{"ghostty", "truecolor"},
		{"iTerm.app", "truecolor"},
		{"WezTerm", "24bit"},
		{"", "truecolor"},
		{"", ""},
	} {
		t.Setenv("TERM_PROGRAM", tc.prog)
		t.Setenv("COLORTERM", tc.colorterm)
		if got := DetectProfile(); got != Profile256 {
			t.Errorf("TERM_PROGRAM=%q COLORTERM=%q: profile %s, want 256", tc.prog, tc.colorterm, got)
		}
	}
	t.Setenv("XSCAPES_COLOR", "truecolor")
	if got := DetectProfile(); got != ProfileTrueColor {
		t.Errorf("XSCAPES_COLOR=truecolor: profile %s, want truecolor", got)
	}
}
