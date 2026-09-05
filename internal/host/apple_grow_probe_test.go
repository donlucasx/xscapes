package host

import (
	"fmt"
	"strings"
	"testing"
)

// Terminal.app's rules through big grows, drags and round trips: no scape-
// coloured row may be left in the agent's band. Same shape as the Ghostty
// test; the Terminal.app tests had no such assertion for a grow of more than
// a few rows.
//
// Written 2026-09-05 as an instrument: his Terminal.app day screenshot at
// 120x61 held the night scape's top two rows (sky greys, a pink moon tone
// painted only 20:50-21:50 and 01:55-02:50) at window rows 28-29, inside
// Claude's blank area, where the band's top sits in a 45-row window. GREEN on
// first run in every geometry here, so the host's own sequences do not leave
// them under this model of the terminal; whatever left them is a real
// terminal behaviour the model does not know, or an event outside a resize.
// A trace of a session that shows it is the next instrument.
func TestAppleTerminalResizeLeavesNoScapeRowInTheBand(t *testing.T) {
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
	for _, tc := range []struct {
		name  string
		start int
		steps []int
	}{
		{"grow seventeen rows at once", 44, []int{61}},
		{"grow 44 to 61 in ticks of four", 44, []int{48, 52, 56, 60, 61}},
		{"grow 44 to 61 one row at a time", 44, func() []int {
			var s []int
			for h := 45; h <= 61; h++ {
				s = append(s, h)
			}
			return s
		}()},
		{"shrink eleven then grow back", 61, []int{50, 61}},
		{"shrink two, grow two", 61, []int{59, 61}},
		{"grow, shrink, grow", 44, []int{61, 50, 61}},
		{"a drag down then up, ticks of three", 61, []int{58, 55, 52, 55, 58, 61}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc := newScreen(w, tc.start)
			agent, _ := Band(tc.start)
			sc.feed(Open(true, agent, tc.start))
			drawAgentBox(sc, agent)
			paintScape(sc, agent, tc.start)
			rows, band := tc.start, agent
			for _, h := range tc.steps {
				sc.resizeAlt(w, h)
				next, _ := Band(h)
				sc.feed(resizeSequence(true, RulesFor("Apple_Terminal"), rows, h, band, next))
				rows, band = h, next
				paintScape(sc, band, rows)
			}
			var scapeInBand []int
			for y := 0; y < band; y++ {
				if bg, uniform := sc.bgRunAt(y); uniform && bg == scapeBG {
					scapeInBand = append(scapeInBand, y+1)
				}
			}
			if len(scapeInBand) > 0 {
				t.Errorf("band rows %v (of 1..%d in %d) are painted in the scape's colour", scapeInBand, band, rows)
			}
		})
	}
}
