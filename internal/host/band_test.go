package host

import "testing"

// Band splits the window between the hosted agent and the scape.
//
// The agent's band has to be anchored at row 1: lines that scroll out of a
// DECSTBM region only reach the scrollback when the region starts at the top
// of the screen (measured in Terminal.app -- rows 1-10 keep every scrolled
// line, rows 5-14 keep none). So the scape gets the rows below, and the split
// is the only free parameter.

func TestBandAlwaysSpendsTheWholeWindow(t *testing.T) {
	for h := 4; h <= 120; h++ {
		a, s := Band(h)
		if a+s != h {
			t.Fatalf("h=%d: agent %d + scape %d = %d, want %d", h, a, s, a+s, h)
		}
		if a < 0 || s < 0 {
			t.Fatalf("h=%d: negative rows: agent %d scape %d", h, a, s)
		}
	}
}

// The agent is the thing being used. A scape that squeezes Claude's input box
// off the screen has got the priority backwards.
func TestAgentKeepsAWorkableBandWheneverTheScapeShows(t *testing.T) {
	for h := 4; h <= 120; h++ {
		a, s := Band(h)
		if s > 0 && a < MinAgentRows {
			t.Errorf("h=%d: agent got %d rows with a %d-row scape, want at least %d", h, a, s, MinAgentRows)
		}
	}
}

// Shore stops drawing below six rows and needs a few more than that before the
// beach is worth having. Rather than paint a broken scape, show none.
func TestScapeIsEitherTallEnoughToReadOrAbsent(t *testing.T) {
	for h := 4; h <= 120; h++ {
		_, s := Band(h)
		if s != 0 && s < MinScapeRows {
			t.Errorf("h=%d: scape got %d rows, want 0 or at least %d", h, s, MinScapeRows)
		}
	}
}

// His ruling: "the taller the window the more sand below where we can see what
// claude is working on". Growing the window must never shrink the scape.
func TestScapeNeverShrinksAsTheWindowGrows(t *testing.T) {
	prev := 0
	for h := 14; h <= 120; h++ {
		_, s := Band(h)
		if s < prev {
			t.Errorf("h=%d: scape fell from %d to %d", h, prev, s)
		}
		prev = s
	}
}

// Past a point more beach stops buying anything and the agent should have the
// room instead.
func TestScapeStopsGrowingAtItsCap(t *testing.T) {
	_, s := Band(200)
	if s != MaxScapeRows {
		t.Errorf("at 200 rows the scape took %d, want the %d-row cap", s, MaxScapeRows)
	}
}

func TestShortWindowDropsTheScapeRatherThanTheAgent(t *testing.T) {
	a, s := Band(12)
	if s != 0 {
		t.Errorf("at 12 rows the scape took %d rows, want none", s)
	}
	if a != 12 {
		t.Errorf("agent got %d of 12 rows, want all of them", a)
	}
}

func TestTypicalWindows(t *testing.T) {
	for _, c := range []struct{ h, agent, scape int }{
		{24, 15, 9},
		{30, 18, 12},
		{52, 32, 20},
		{79, 51, 28},
	} {
		a, s := Band(c.h)
		if a != c.agent || s != c.scape {
			t.Errorf("Band(%d) = %d/%d, want %d/%d", c.h, a, s, c.agent, c.scape)
		}
	}
}

// BandWith lets the split be set by hand: he asked for more scape than two
// fifths gives, and the right number is a matter of taste, not measurement.

func TestBandWithZeroIsTheAutomaticSplit(t *testing.T) {
	wa, ws := BandWith(51, 0)
	a, sc := Band(51)
	if wa != a || ws != sc {
		t.Errorf("BandWith(51,0) = %d/%d, want the automatic %d/%d", wa, ws, a, sc)
	}
}

func TestBandWithHonoursTheRequestedRows(t *testing.T) {
	a, s := BandWith(51, 30)
	if s != 30 || a != 21 {
		t.Errorf("BandWith(51,30) = %d/%d, want 21/30", a, s)
	}
}

// A request that would squeeze the agent out is capped, not obeyed. The agent
// is the thing being used.
func TestBandWithNeverStarvesTheAgent(t *testing.T) {
	a, s := BandWith(51, 48)
	if a < MinAgentRows {
		t.Errorf("BandWith(51,48) left the agent %d rows, want at least %d", a, MinAgentRows)
	}
	if a+s != 51 {
		t.Errorf("rows do not sum: %d + %d != 51", a, s)
	}
}

func TestBandWithRaisesATinyRequestToSomethingReadable(t *testing.T) {
	_, s := BandWith(51, 3)
	if s != MinScapeRows {
		t.Errorf("BandWith(51,3) gave the scape %d rows, want %d", s, MinScapeRows)
	}
}
