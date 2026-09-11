package main

import "fmt"

// verdict is the ranked list. Each line is one reason, and the reason is a
// measurement from above rather than a preference.
func verdict() {
	fmt.Println("\n\n==== 10. RANKED, WITH THE MEASUREMENT THAT RANKS IT ====")
	for _, v := range [][3]string{
		{"1", "A5  ARRIVAL FLARE (fade white -> the star's own tone, ~1s)",
			"Passes all three tests and costs one cell. Peak sits +47.6 luma over the\n      settled star and +155.6 over the sky, so it is unmissable; the count is\n      never wrong because it is the same cell, the same glyph, the whole time;\n      and at his measured p50 star gap of 5.4 min it is on screen 0.3% of frames.\n      It is also the direct answer to the thing he has said twice -- that he\n      never noticed this channel."},
		{"2", "A4  ARRIVAL AS A FALL (the star streaks into its own place), NO TAIL",
			"Same three passes, and it reads as a shooting star without binding one to\n      an event that never fires. Every frame of the fall carries exactly 12 '*'\n      cells. ⚠ Ship it WITHOUT a trail: the faint glyphs a trail needs are the\n      ambient field's own ('.', 0xb7, '+'), so a tail either reads as dust or\n      adds three more marks. Strictly more work than A5 for a similar effect --\n      pick it only if he wants the movement."},
		{"3", "D2  AGE BY HUE  -- only if he asks for it after seeing A",
			"The rule half-fails it (recency is a second variable on a counting\n      channel) and the cube finishes it off: 6 distinct entries for 26 stars,\n      with 46 luma between oldest and newest, which is a brightness gradient in\n      disguise heading for the 0.85 floor."},
		{"x", "C   MAGNITUDE / a tally in fives",
			"REJECT as a TALLY. The partition is [9 1 1 1 9] at HEAD and [3 4 4 4 6] on\n      the tree's blue noise, and even the good one does not read: five marks make\n      a tally when they TOUCH, not when a Voronoi cell says so. ⚠ SEPARATE FROM\n      THIS: magnitude as TEXTURE is already in the working tree (todoStarMag) and\n      my rejection does not reach it -- it never asks the eye to group. Its open\n      question is the ring's and it is his: is a second weight a second meaning?"},
		{"x", "B   CONSTELLATION LINES",
			"REJECT. At 26 stars the strokes fill 101% of the sky's cells, cross each\n      other 352 times, pass within one cell of a star 252 times and put 34 cells\n      on the moon's face. On the tree's blue noise it is better and still\n      hopeless: 843 cells, 61%, 248 crossings, 28 on the disc. No placement saves\n      it -- short lines need consecutive stars to be neighbours, and both\n      placements exist to keep them apart."},
		{"x", "E   SHOOTING STAR ON A RARE EVENT",
			"REJECT, counted not remembered: compact fires 0 times in 76,495 events\n      across 123 spools, and every event that DOES fire already owns a channel\n      (needs_input/done = the bubble, error = the companion)."},
		{"x", "K   THE NEWEST STAR BREATHES",
			"REJECT. Which star is newest would live only in motion, and a screenshot\n      has no motion. It also looks exactly like the defect he reported on 09-10:\n      \"once they appear, they should not disappear IMO.\""},
	} {
		fmt.Printf("\n  %s. %s\n      %s\n", v[0], v[1], v[2])
	}
	fmt.Println("\n  ⚠ NONE OF THIS ANSWERS THE FIRST TWO THIRDS OF HIS NOTE. \"Randomly across")
	fmt.Println("    the sky\" and \"spread them more vertically\" are PLACEMENT, measured in")
	fmt.Println("    notes/s28-starroom: the band is rows 1..6 of 11, 55% of the sky and all of")
	fmt.Println("    it the top half. A flare on a constellation still crammed into six rows is")
	fmt.Println("    a brighter version of the thing he is complaining about. Placement first.")
}
