package host

// The window is split in two: the agent's band on top, the scape below it.
//
// Top is not a preference, it is forced. Lines that scroll out of a DECSTBM
// region reach the scrollback only when the region starts at row 1 -- measured
// in Terminal.app, region 1-10 keeps every scrolled line and region 5-14 keeps
// none (notes/claude-terminal-emissions.md). Putting anything above the agent
// would cost the user the ability to scroll back through its output, which is
// far too much to pay for a strip of sky.
//
// So the scape reads downward from the agent instead of upward from the
// bottom: a strip of sky under the band, then the sea, then the beach. A
// horizon works just as well with the agent above it.
const (
	// MinAgentRows is the shortest band Claude Code is usable in: its input
	// box, its status lines, and a few rows of answer.
	MinAgentRows = 10
	// MinScapeRows is the shortest scape worth painting. Shore stops drawing
	// below six rows, and wants a couple more before the beach can hold a
	// line of work history.
	MinScapeRows = 8
	// MaxScapeRows caps the scape. Past twenty-eight rows more beach stops
	// buying anything and the agent should have the room instead.
	MaxScapeRows = 28
)

// Band divides h rows between the agent and the scape. They always sum to h.
//
// Nine twentieths of the window: 13 rows of scape at 30, 23 at 52, the 28-row
// cap from 63 up. It was a third, then two fifths, and it keeps moving the same
// way -- "lets make it a tad taller so all the xscape layers can shine". A short
// scape spends its rows on the beach floor and leaves the sea a strip; the sky,
// the sea and the beach only read as three things once there are rows for them. Never the majority, though: the
// agent is what is being used. A window too short to hold both gets no scape
// rather than a broken one.
func Band(h int) (agent, scape int) {
	if h < MinAgentRows+MinScapeRows {
		return h, 0
	}
	s := h * 9 / 20
	if s < MinScapeRows {
		s = MinScapeRows
	}
	if s > MaxScapeRows {
		s = MaxScapeRows
	}
	if h-s < MinAgentRows {
		s = h - MinAgentRows
	}
	return h - s, s
}

// BandWith is Band with the scape's rows set by hand. Zero means automatic.
//
// The right number here is taste, not measurement: how much shoreline you want
// beside your work is a preference, so -scape exists rather than another
// constant argued into place.
func BandWith(h, want int) (agent, scape int) {
	if want <= 0 {
		return Band(h)
	}
	if h < MinAgentRows+MinScapeRows {
		return h, 0
	}
	if want < MinScapeRows {
		want = MinScapeRows
	}
	if h-want < MinAgentRows {
		want = h - MinAgentRows
	}
	return h - want, want
}
