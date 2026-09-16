package scape

import "testing"

// TestFormatTokens: the spend reads in at most four characters at every
// magnitude a session can reach, so the counter's width is known.
func TestFormatTokens(t *testing.T) {
	cases := map[int64]string{
		0: "0", 950: "950", 999: "999", 1000: "1.0k", 9800: "9.8k", 9970: "10k",
		12_345: "12k", 845_000: "845k", 999_499: "999k", 1_000_000: "1.0M",
		1_234_567: "1.2M", 9_960_000: "10M", 12_000_000: "12M", 999_000_000: "999M",
		999_999: "1.0M", 999_999_999: "1.0B",
	}
	for n, want := range cases {
		if got := FormatTokens(n); got != want {
			t.Errorf("FormatTokens(%d) = %q, want %q", n, got, want)
		}
	}
	for n := int64(1); n < 1_000_000_000; n = n*10 + 7 {
		if got := FormatTokens(n); len(got) > 4 {
			t.Errorf("FormatTokens(%d) = %q, five characters", n, got)
		}
	}
	// The label ends one cell in from the right edge, on row one, and its
	// cells are inside the reserved ground at every width from the floor up;
	// under the floor there is no label and no ground.
	for w := 20; w < TokensMinWidth; w++ {
		if _, _, _, ok := TokensLabel(w, 1_234_567); ok {
			t.Errorf("width %d: a label under the floor", w)
		}
		if x0, x1, _ := TokensGround(w); x0 <= x1 {
			t.Errorf("width %d: ground reserved under the floor", w)
		}
	}
	for w := TokensMinWidth; w <= 200; w++ {
		for _, n := range []int64{1, 9999, 999_999, 9_999_999, 999_000_000} {
			txt, x, y, ok := TokensLabel(w, n)
			if !ok {
				t.Fatalf("width %d, %d tokens: no label", w, n)
			}
			if x+len(txt)-1 != w-1-TokensMargin || y != TokensRow {
				t.Errorf("width %d: label %q at (%d,%d) does not end on column %d", w, txt, x, y, w-1-TokensMargin)
			}
			gx0, gx1, gy := TokensGround(w)
			if x < gx0 || x+len(txt)-1 > gx1 || y != gy {
				t.Errorf("width %d: label %q at %d..%d is outside the ground %d..%d", w, txt, x, x+len(txt)-1, gx0, gx1)
			}
		}
	}
	if _, _, _, ok := TokensLabel(80, 0); ok {
		t.Error("no reading, yet a label")
	}
}

// TestNoPlacedStarSitsUnderTheSpendCounter: at his geometries, every seed
// and every count, no lit star is on a cell the counter can be painted on
// (TokensGround). A dart that lands there is re-thrown; the rest of the
// constellation is where it always was (the golden in ambient_test holds).
func TestNoPlacedStarSitsUnderTheSpendCounter(t *testing.T) {
	checked := 0
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 11, 23, 42} {
			for done := 1; done <= 32; done++ {
				_, c := starFrame(g.w, g.h, seed, done, 20.0/24, 0.3)
				for _, p := range litStars(c) {
					checked++
					if inTokensGround(g.w, p) {
						t.Errorf("%dx%d seed %d done %d: the star at %v is under the spend counter", g.w, g.h, seed, done, p)
					}
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no stars checked")
	}
}
