package main

import "fmt"

// paceSpan / paceAt / stillFor are COPIED from pace.go, not imported: pace.go
// lives in package main at the repo root and there is no way to reach it from
// here. If pace.go changes, this copy is stale.
func paceSpan(w int) int {
	s := w / 16
	if s > 6 {
		s = 6
	}
	if s < 1 {
		s = 1
	}
	return s
}

func tri(n, span int) float64 {
	if span <= 0 {
		return 0
	}
	p := n % (2 * span)
	if p < 0 {
		p += 2 * span
	}
	if p > span {
		p = 2*span - p
	}
	return float64(p)
}

// paceSnap is the thing any approach has to deal with first, and it is already
// shipped: stillFor() holds the companion still at NeedsYou, and "still" is
// implemented as dx = 0 -- HOME, which is the column nearest the frame edge.
//
// So at the instant the ask fires, the companion does not stop where it was
// standing. It TELEPORTS outward, by however far it had paced, in one frame.
// Away from the reader, at the exact moment it wants the reader.
func paceSnap() {
	fmt.Printf("\n%-9s %6s %8s  %s\n", "geom", "span", "worst", "where it is standing when the ask lands (- is inward)")
	for _, w := range []int{40, 80, 111, 124, 143, 153} {
		span := paceSpan(w)
		var seq string
		for n := 0; n <= 2*span; n++ {
			seq += fmt.Sprintf("%d ", -int(tri(n, span)))
		}
		fmt.Printf("%-9d %6d %8s  %s-> 0 the frame the pose flips to NeedsYou\n", w, span,
			fmt.Sprintf("%d cells", span), seq)
	}
	fmt.Println("The jump is instant and it is outward. An approach cannot be added on top of")
	fmt.Println("this; the snap has to be replaced by a walk, whichever direction wins.")
}
