package term

import "testing"

// The scene must look the same in every terminal profile.
//
// Indices 0-15 are the ANSI slots, and those ARE user-editable -- they are the
// sixteen swatches in Terminal.app's profile editor, and they differ between
// Solid Colors, Grass, Man Page and the rest. Everything from 16 up is fixed by
// the xterm-256 standard and cannot be themed. So the whole promise of "the
// same picture on every profile" rests on never emitting an index below 16.
//
// Exhaustive over all 16.7M colours, because "by construction" is a claim about
// code I wrote and this is a claim about output.
func TestNeverEmitsAThemeableColour(t *testing.T) {
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				if i := (RGB{uint8(r), uint8(g), uint8(b)}).Index256(); i < 16 || i > 255 {
					t.Fatalf("rgb(%d,%d,%d) -> index %d, which the user's profile can repaint", r, g, b, i)
				}
			}
		}
	}
}
