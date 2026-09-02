package host

import "testing"

// The scape repaints every cell of every row, every frame: measured at ~30KB a
// frame, and at 20fps that is 0.6MB/s of fully-coloured cells poured into a
// terminal that is already rendering the agent. Most of those cells are
// identical to the frame before -- the sky and most of the beach do not move.
// Rows that did not change should not be sent at all.

func TestFirstFrameRepaintsEveryRow(t *testing.T) {
	var d damage
	got := d.changed([]string{"a", "b", "c"})
	if len(got) != 3 {
		t.Errorf("first frame repainted %v, want all three rows", got)
	}
}

func TestUnchangedRowsAreNotRepainted(t *testing.T) {
	var d damage
	d.changed([]string{"sky", "sea", "sand"})
	got := d.changed([]string{"sky", "SEA", "sand"})
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("repainted %v, want only row 1", got)
	}
}

func TestAnUnchangedFrameSendsNothing(t *testing.T) {
	var d damage
	rows := []string{"sky", "sea", "sand"}
	d.changed(rows)
	if got := d.changed(rows); len(got) != 0 {
		t.Errorf("repainted %v on an identical frame, want nothing", got)
	}
}

// The trap in every damage-tracked renderer: the screen is cleared or the
// window resized underneath the tracker, and it happily reports "nothing
// changed" onto rows that are now blank. Anything that touches the screen
// behind its back must reset it.
func TestResetForcesAFullRepaintEvenIfNothingChanged(t *testing.T) {
	var d damage
	rows := []string{"sky", "sea", "sand"}
	d.changed(rows)
	d.reset()
	if got := d.changed(rows); len(got) != 3 {
		t.Errorf("repainted %v after a reset, want all three rows", got)
	}
}

func TestADifferentRowCountRepaintsEverything(t *testing.T) {
	var d damage
	d.changed([]string{"sky", "sea", "sand"})
	got := d.changed([]string{"sky", "sea", "sand", "more"})
	if len(got) != 4 {
		t.Errorf("repainted %v when the row count changed, want all four", got)
	}
}

// Damage tracking has one failure mode that never heals: if a cell on screen
// stops matching what the tracker believes is there -- anything at all wrote
// over it -- that row is skipped for as long as its content does not change,
// and the wrong pixels stay wrong. The sky and the beach can sit unchanged for
// minutes at a time, so "until it changes" can mean "forever".
//
// So the tracker forgets everything periodically and repaints in full.
func TestDamageForgetsPeriodicallySoStaleCellsHeal(t *testing.T) {
	var d damage
	rows := []string{"sky", "sea", "sand"}
	full := 0
	for i := 0; i < 200; i++ {
		if len(d.changed(rows)) == len(rows) {
			full++
		}
	}
	if full < 2 {
		t.Errorf("in 200 identical frames the tracker repainted in full %d times: stale cells would never heal", full)
	}
	if full > 20 {
		t.Errorf("repainted in full %d times in 200 frames: that is most of what damage tracking was meant to save", full)
	}
}
