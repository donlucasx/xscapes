package host

// damage tracks which of the scape's rows changed since the last frame, so a
// frame sends only what moved.
//
// Measured before it existed: a frame was ~30KB and at 20fps that is 0.6MB/s of
// fully-coloured cells, every one of them rewritten whether or not it had
// changed, into a terminal that is also rendering the agent. Most of a scape
// does not move between frames -- the sky sits still, the beach only changes
// when a line is written into it -- so most of that traffic was redundant.
//
// Row granularity rather than cell: a row is one string that Render already
// produced, comparing them is a string compare, and the rows that do move (the
// sea) move along their whole length anyway.
type damage struct {
	prev []string
}

// changed returns the indices of the rows that need painting.
func (d *damage) changed(rows []string) []int {
	full := len(d.prev) != len(rows)
	out := make([]int, 0, len(rows))
	for i, r := range rows {
		if full || d.prev[i] != r {
			out = append(out, i)
		}
	}
	d.prev = append(d.prev[:0:0], rows...)
	return out
}

// reset forces the next frame to repaint everything.
//
// This is the whole safety of the thing. Anything that touches the screen
// behind the tracker's back -- a resize, a clear, re-establishing the band --
// leaves it believing rows are painted that are now blank, and it would
// cheerfully skip them forever.
func (d *damage) reset() { d.prev = nil }
