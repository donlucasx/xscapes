package host

import "testing"

// Terminal.app's alternate screen on a WIDTH change, measured 2026-09-05
// (notes/width-audit.md): rows keep every cell at the widest the window has
// been, a narrower window clips, a wider one shows the retained cells again,
// and an erase-line reaches the visible width only. Under the default rule a
// row is cut and padded, which is what the model assumed before.
func TestTerminalAppRetainsCellsBeyondTheWidth(t *testing.T) {
	for _, retain := range []bool{true, false} {
		sc := newScreen(20, 3)
		sc.retainWidth = retain
		sc.feed("\x1b[1;1HAAAAAAAAAAAAAAAAAAA<\x1b[2;1HBBBBBBBBBBBBBBBBBBB<")
		sc.resizeAlt(12, 3)
		sc.feed("\x1b[2;1H\x1b[2K") // erase row 2 while narrow
		sc.resizeAlt(20, 3)
		r1, r2 := sc.rowAt(0), sc.rowAt(1)
		if retain {
			if r1 != "AAAAAAAAAAAAAAAAAAA<" {
				t.Errorf("retain: row 1 after narrow and widen = %q, want the whole row back", r1)
			}
			if r2 != "            BBBBBBB<" {
				t.Errorf("retain: erased row 2 = %q, want 12 blanks then the retained tail", r2)
			}
		} else {
			// rowAt trims trailing blanks, so the padding is not visible here.
			if r1 != "AAAAAAAAAAAA" {
				t.Errorf("cut: row 1 = %q, want cut to 12", r1)
			}
			if r2 != "" {
				t.Errorf("cut: erased row 2 = %q, want blank", r2)
			}
		}
	}
}
