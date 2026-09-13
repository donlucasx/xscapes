package main

import (
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/companion"
)

// The muzzle fill must be INVISIBLE at the shipped size.
//
// His ruling of 2026-09-12 removed a mouth that only exists at 24x14: the
// shipped cat carries a gap at CatBody row 11 which ToQuadrant's row-pair OR
// swallows, and doubling puts it on a clean boundary where it survives. Filling
// row 11 is therefore free at 12x7 and load-bearing at 24x14 -- but "free" is a
// claim about the RENDERED frame, so it gets asserted rather than argued.
func TestFillingTheMuzzleIsInvisibleAtTheShippedSize(t *testing.T) {
	a := companion.ParseBitmap(companion.CatBody).ToQuadrant()
	b := companion.ParseBitmap(catBodyNoMouth).ToQuadrant()
	if strings.Join(a, "\n") != strings.Join(b, "\n") {
		t.Fatalf("filling the muzzle changed the 12x7 render:\nwas  %q\nnow  %q", a, b)
	}
	// A positive control: the two sources really are different, so the test
	// above is comparing something. Without this it would also pass if both
	// sides were accidentally the same bitmap.
	same := 0
	for i := range companion.CatBody {
		if companion.CatBody[i] == catBodyNoMouth[i] {
			same++
		}
	}
	if same == len(companion.CatBody) {
		t.Fatal("catBodyNoMouth is identical to CatBody -- the fill is missing, " +
			"so the invisibility check above proves nothing")
	}
}

// And it must be VISIBLE at 24x14, or it is not a fix.
func TestFillingTheMuzzleRemovesTheMouthAt2x(t *testing.T) {
	with := companion.ParseBitmap(double2x(companion.CatBody)).ToQuadrant()
	without := companion.ParseBitmap(double2x(catBodyNoMouth)).ToQuadrant()
	diff := 0
	for i := range with {
		if with[i] != without[i] {
			diff++
		}
	}
	if diff != 1 {
		t.Errorf("doubled render differs on %d cell rows, want exactly 1 (the jaw)", diff)
	}
	// The mouth is the lower half of the centre cells going empty, which
	// renders as the upper-half glyph. It must be gone.
	if strings.Contains(without[5], "▀▀▀") {
		t.Errorf("row 5 still carries the mouth: %q", without[5])
	}
	if !strings.Contains(with[5], "▀▀▀") {
		t.Errorf("the control is wrong: the unfixed row 5 should carry the mouth, got %q", with[5])
	}
}

// The eye sockets are where the art puts them, not where the crab's are.
//
// This is the assertion that would have caught the mockup drawing both eyes on
// the animal's forehead: it passed {8, 14}, which are the CRAB's near-pose eye
// cells, and mirrored they land on ink.
func TestTheDoubledCatsEyeSocketsAreEmpty(t *testing.T) {
	bm := companion.ParseBitmap(double2x(catBodyNoMouth))
	// ⚠ CHECK THE MIRRORED FRAME, not a mirrored INDEX into the unmirrored one.
	// The body occupies cells 1..17 of a 24-cell box -- the right six are the
	// tail's corridor -- so it is NOT centred, and mirroring slides the whole
	// animal across the box. The first version of this test reflected the cell
	// index instead of the bitmap and failed on correct art.
	for _, arm := range []struct {
		name string
		q    []string
	}{{"unmirrored", bm.ToQuadrant()}, {"mirrored", bm.Mirrored().ToQuadrant()}} {
		row := []rune(arm.q[catRung2EyeRow])
		w := len(row)
		for _, s := range catRung2Sockets {
			for cx := s[0]; cx <= s[1]; cx++ {
				x := cx
				if arm.name == "mirrored" {
					x = w - 1 - cx
				}
				if row[x] != ' ' {
					t.Errorf("%s: socket cell %d on row %d is %q, want empty",
						arm.name, x, catRung2EyeRow, row[x])
				}
			}
		}
		// The crab's cells, as the negative control: they must NOT be empty,
		// which is exactly why borrowing them was a bug.
		empty := 0
		for _, cx := range []int{8, 14} {
			x := cx
			if arm.name == "mirrored" {
				x = w - 1 - cx
			}
			if row[x] == ' ' {
				empty++
			}
		}
		if empty == 2 {
			t.Errorf("%s: both of the crab's eye cells are empty here, so borrowing them "+
				"would have been harmless and this test does not measure the defect it was "+
				"written for", arm.name)
		}
	}
}
