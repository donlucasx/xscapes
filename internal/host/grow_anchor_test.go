package host

import (
	"fmt"
	"strings"
	"testing"
)

// A taller window brings back the rows that scrolled off the band and keeps
// the agent's text on the band's bottom, as a plain terminal's main screen
// does. (With no history nothing moves; the older tests cover that.)
//
// His report of 2026-11-08: `xscapes hermes`, window made taller, and Hermes's
// prompt stayed where it was with ten blank rows under it (Terminal.app trace
// 20261108-130027, replayed through this model: chrome on rows 13-16 of a
// 26-row band). Plain Hermes in the same terminals kept its prompt at the
// bottom, because a main-screen grow pulls scrollback in from the top and the
// cursor goes down with the text. Hermes (prompt_toolkit) redraws only its
// bottom chrome, by relative moves from the cursor, and never repaints the
// rows above, so whatever the grow leaves is what stays.
//
// The host used to keep the text at the TOP on a grow (Terminal.app's push
// scrolled back up, Ghostty's top-anchored grow left alone). Claude Code hid
// it: it repaints its whole band after a resize. A shrink already anchors to
// the bottom (RebindShrinkAlt), so this makes the two directions agree.
func TestGrowKeepsTheTextOnTheBandsBottom(t *testing.T) {
	const w = 120
	const scapeBG = 25
	paintScape := func(sc *screen, agent, rows int) {
		var b strings.Builder
		b.WriteString(BeginPaint())
		for r := agent + 1; r <= rows; r++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[48;5;%dm%s\x1b[0m", r, scapeBG, strings.Repeat(" ", w))
		}
		b.WriteString(EndPaint(agent))
		sc.feed(b.String())
	}
	terminals := []struct {
		name   string
		rules  Rules
		resize func(sc *screen, w, h int)
		setup  func(sc *screen)
	}{
		{"Terminal.app", AppleTerminalRules, (*screen).resizeAlt, func(*screen) {}},
		{"Ghostty", XTermRules, (*screen).resizeGhostty, func(sc *screen) { sc.restoreAbsolute = true }},
	}
	for _, term := range terminals {
		for _, tc := range []struct {
			name  string
			start int
			steps []int
		}{
			{"grow one row", 30, []int{31}},
			{"grow ten rows", 30, []int{40}},
			{"a drag up, ticks of two", 30, []int{32, 34, 36}},
			{"shrink then grow back", 36, []int{30, 36}},
		} {
			t.Run(term.name+"/"+tc.name, func(t *testing.T) {
				sc := newScreen(w, tc.start)
				term.setup(sc)
				agent, _ := Band(tc.start)
				sc.feed(Open(true, agent, tc.start))
				drawAgentBox(sc, agent)
				paintScape(sc, agent, tc.start)
				// Forty rows have scrolled off the band; a grow brings back the
				// newest, as Host.pullBack hands them over.
				var history []string
				for i := 1; i <= 40; i++ {
					history = append(history, fmt.Sprintf("HISTORY ROW %02d", i))
				}
				rows, band := tc.start, agent
				for _, h := range tc.steps {
					term.resize(sc, w, h)
					next, _ := Band(h)
					var back []string
					if g := next - band; h > rows && g > 0 {
						back = history[len(history)-g:]
						history = history[:len(history)-g]
					}
					sc.feed(resizeSequence(true, term.rules, rows, h, band, next, back...))
					rows, band = h, next
					paintScape(sc, band, rows)
				}
				// The band's top row is history, newest last, right above
				// the transcript's first row.
				if r := sc.rowAt(0); band > agent && !strings.Contains(r, "HISTORY ROW") {
					t.Errorf("row 1 = %q, want a history row brought back", strings.TrimSpace(r))
				}

				// Before any redraw: the old box sits on the band's last rows
				// and the cursor is still on its input row.
				if r := sc.rowAt(band - 3); !strings.Contains(r, "old box") {
					t.Errorf("row %d (band 1..%d) = %q, want the old box's top: the text did not follow the band's bottom", band-2, band, strings.TrimSpace(r))
				}
				if got := sc.y + 1; got != band-1 {
					t.Errorf("cursor on row %d, want %d (the input row)", got, band-1)
				}

				// Hermes redraws its chrome relatively from the cursor.
				sc.feed("\r\x1b[1A\x1b[J+==== new box ====+\r\n| > hi again      |\r\n" +
					"+=================+\r\x1b[1A\x1b[13C")
				var box, scapeInBand []int
				for y := 0; y < sc.h; y++ {
					if strings.Contains(sc.rowAt(y), "new box") {
						box = append(box, y+1)
					}
					if bg, uniform := sc.bgRunAt(y); y < band && uniform && bg == scapeBG {
						scapeInBand = append(scapeInBand, y+1)
					}
				}
				if len(box) != 1 || box[0] != band-2 {
					t.Errorf("new box top at rows %v, want row %d: the chrome is not on the band's bottom", box, band-2)
				}
				if len(scapeInBand) > 0 {
					t.Errorf("band rows %v are painted in the scape's colour", scapeInBand)
				}
				// The transcript's last row stays right above the box.
				if r := sc.rowAt(band - 4); !strings.Contains(r, "TRANSCRIPT ROW") {
					t.Errorf("row %d = %q, want the transcript's last row above the box", band-3, strings.TrimSpace(r))
				}
			})
		}
	}
}
