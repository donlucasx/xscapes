package main

// ladder: which sizes the medium can actually reach between 12x7 and 24x14,
// and which of them is closest to a UNIFORM push-in of the shipped crab.
//
// A push-in preserves aspect. Source pixels are square on screen, so the test
// is (w/24) vs (h/28). Every draft picked a size and then argued the room; this
// picks the size the ROOM allows that is closest to a real magnification, and
// draws it so the claim is not arithmetic alone.

import (
	"fmt"
	"math"
	"strings"
)

// scaleSym is symmetric nearest-neighbour: the column map is built for the left
// half and mirrored, so a symmetric source stays symmetric. Asymmetric column
// duplication is what made the naive scale render legs on different rungs.
func scaleSym(rows []string, dw, dh int) []string {
	sw := len([]rune(rows[0]))
	sh := len(rows)
	cm := make([]int, dw)
	for j := 0; j < dw; j++ {
		cm[j] = int(math.Floor((float64(j) + 0.5) * float64(sw) / float64(dw)))
	}
	for j := 0; j < dw/2; j++ {
		cm[dw-1-j] = sw - 1 - cm[j]
	}
	rm := make([]int, dh)
	for i := 0; i < dh; i++ {
		rm[i] = int(math.Floor((float64(i) + 0.5) * float64(sh) / float64(dh)))
	}
	out := make([]string, dh)
	for i := 0; i < dh; i++ {
		src := []rune(rows[rm[i]])
		var sb strings.Builder
		for j := 0; j < dw; j++ {
			sb.WriteRune(src[cm[j]])
		}
		out[i] = sb.String()
	}
	return out
}

func asymRows(rows []string) int {
	n := 0
	for _, r := range rows {
		rr := []rune(r)
		for i := 0; i < len(rr)/2; i++ {
			if rr[i] != rr[len(rr)-1-i] {
				n++
				break
			}
		}
	}
	return n
}

func ladder() {
	fmt.Println("=== 13. EVERY SIZE THE MEDIUM CAN REACH between the shipped crab and a 2x,")
	fmt.Println("    scored on how close it is to a UNIFORM push-in (aspect 24:28 = 0.857).")
	fmt.Println("    'rows' is the cell height, which is what the beach has to pay for.")
	fmt.Println()
	fmt.Printf("%8s %8s %7s %7s %8s %9s  %s\n", "cells", "src px", "xscale", "yscale", "aspect", "off-push", "note")
	type row struct {
		cw, chh int
	}
	for _, s := range []row{{12, 7}, {13, 8}, {14, 8}, {14, 9}, {15, 9}, {16, 9}, {16, 10}, {16, 11}, {18, 10}, {20, 12}, {24, 14}} {
		w, h := s.cw*2, s.chh*4
		xs, ys := float64(w)/24, float64(h)/28
		asp := float64(w) / float64(h)
		off := asp/0.8571428 - 1
		note := ""
		switch {
		case s.cw == 12 && s.chh == 7:
			note = "shipped"
		case s.cw == 14 && s.chh == 9:
			note = "A"
		case s.cw == 16 && s.chh == 11:
			note = "B"
		case s.cw == 24 && s.chh == 14:
			note = "D (exact 2x)"
		}
		if math.Abs(off) < 0.05 && note == "" {
			note = "<- within 5% of a true push-in"
		}
		fmt.Printf("%8s %8s %7.2f %7.2f %8.3f %+8.0f%%  %s\n",
			fmt.Sprintf("%dx%d", s.cw, s.chh), fmt.Sprintf("%dx%d", w, h), xs, ys, asp, off*100, note)
	}

	fmt.Println("\n=== 14. THE THREE CANDIDATE RUNGS, GENERATED from the shipped ask by symmetric")
	fmt.Println("    nearest-neighbour, beside A's hand-drawn 14x9. Mirrored, as shipped.")
	fmt.Println("    A generated rung is a check on the SHAPE, not a proposal to ship pixels.")
	fmt.Println()
	ship := join(crabAskRows(), crabLowerRows())
	var cols [][]string
	var heads []string
	for _, s := range []struct {
		cw, chh int
		lab     string
	}{{12, 7, "shipped 12x7"}, {14, 8, "gen 14x8"}, {15, 9, "gen 15x9"}, {16, 9, "gen 16x9"}} {
		r := scaleSym(ship, s.cw*2, s.chh*4)
		cols = append(cols, renderM(r))
		heads = append(heads, fmt.Sprintf("%s a%d", s.lab, asymRows(r)))
	}
	cols = append(cols, renderM(join(aUpper, aLower)))
	heads = append(heads, "A 14x9 (hand)")
	fmt.Println(sideBySide(cols, heads, 3, true))

	fmt.Println("    'a<N>' is how many of the scaled source rows came out ASYMMETRIC.")
	fmt.Println("    Zero means the scale did not put the legs on different rungs.")
	fmt.Println()
	for _, s := range []struct{ cw, chh int }{{14, 8}, {15, 9}, {16, 9}, {16, 10}} {
		r := scaleSym(ship, s.cw*2, s.chh*4)
		q := render(r)
		fmt.Printf("    gen %dx%-3d asym rows %d   ink %3d  edge %3d (%2.0f%%)  %s\n",
			s.cw, s.chh, asymRows(r), inkCells(q), edgeCells(q),
			100*float64(edgeCells(q))/float64(inkCells(q)), glyphHist(q))
	}
	q := render(join(aUpper, aLower))
	fmt.Printf("    A 14x9      (hand)        ink %3d  edge %3d (%2.0f%%)  %s\n",
		inkCells(q), edgeCells(q), 100*float64(edgeCells(q))/float64(inkCells(q)), glyphHist(q))
	q = render(ship)
	fmt.Printf("    shipped 12x7              ink %3d  edge %3d (%2.0f%%)  %s\n",
		inkCells(q), edgeCells(q), 100*float64(edgeCells(q))/float64(inkCells(q)), glyphHist(q))

	fmt.Println("\n=== 15. IS crabDone DISTINGUISHABLE FROM crabWork ONCE RENDERED?")
	w1 := render(join(crabWorkRows(), crabLowerRows()))
	d1 := render(join(crabDoneRows(), crabLowerRows()))
	diff := 0
	for i := range w1 {
		a, b := []rune(w1[i]), []rune(d1[i])
		for j := range a {
			if a[j] != b[j] {
				diff++
			}
		}
	}
	fmt.Printf("    crabWork vs crabDone: %d of %d cells differ.\n", diff, len(w1)*len([]rune(w1[0])))
	fmt.Println(sideBySide([][]string{w1, d1, render(join(crabAskRows(), crabLowerRows()))},
		[]string{"work", "done", "ask"}, 3, false))
}

func pickShow() {
	fmt.Println("\n=== 16. THE PICK, AUTHORED AND RENDERED: 16 cells x 9 rows (32x36 px).")
	fmt.Println("    Eye cells {5,9}, cell rows 1-2, 2 cells wide by 2 tall.")
	fmt.Println()
	rows := join(pickUpperRows(), pickLower)
	q := render(rows)
	qm := renderM(rows)
	eq := render(pickEye)
	var mk []string
	for _, r := range eq {
		mk = append(mk, strings.Map(func(x rune) rune {
			if x == ' ' {
				return ' '
			}
			return 'E'
		}, r))
	}
	withEye := stamp(stamp(qm, eq, 5, 1), eq, 9, 1)
	withMark := stamp(stamp(qm, mk, 5, 1), mk, 9, 1)
	fmt.Println(sideBySide([][]string{qm, withEye, withMark},
		[]string{"body (mirrored)", "+ eye pass", "+ eye as E"}, 3, false))
	fmt.Printf("    eye alone: %v\n", eq)
	w, h := srcBBox(rows)
	fmt.Printf("    src bbox %dx%d px  aspect %.3f (shipped 0.857, off by %+.0f%%)  ink %d cells  edge %d (%.0f%%)\n",
		w, h, float64(w)/float64(h), (float64(w)/float64(h)/0.8571428-1)*100,
		inkCells(q), edgeCells(q), 100*float64(edgeCells(q))/float64(inkCells(q)))
	fmt.Printf("    %s\n", glyphHist(q))

	fmt.Println("\n    BOTTOM-ALIGNED beside the shipped crab and A's 14x9:")
	fmt.Println()
	ship := renderM(join(crabAskRows(), crabLowerRows()))
	ship = stamp(stamp(ship, []string{"O"}, 12-1-4, 2), []string{"O"}, 12-1-7, 2)
	aq := renderM(join(aUpper, aLower))
	ae := render(aEye)
	aq = stamp(stamp(aq, ae, 14-2-4, 2), ae, 14-2-8, 2)
	fmt.Println(sideBySide([][]string{ship, aq, withEye},
		[]string{"shipped 12x7", "A 14x9", "PICK 16x9"}, 4, true))

	fmt.Println("    TOP-ALIGNED -- which is how they are actually drawn: same top row,")
	fmt.Println("    the body growing DOWN into the two rows already under its feet.")
	fmt.Println()
	fmt.Println(sideBySide([][]string{ship, aq, withEye},
		[]string{"shipped 12x7", "A 14x9", "PICK 16x9"}, 4, false))

	fmt.Println("    THE HEAD COLUMN. An even-width sprite drawn at catX-(W-12)/2 keeps the")
	fmt.Println("    shipped centre exactly, so bubbleX needs no change:")
	for _, W := range []int{12, 14, 16} {
		off := (W - 12) / 2
		fmt.Printf("        W=%2d drawn at catX-%d: box catX%+d..catX%+d, centre catX%+.1f\n",
			W, off, -off, W-1-off, float64(-off)+float64(W-1)/2)
	}
}

func eyeVocab() {
	fmt.Println("\n=== 18. THE EYE AT 2 CELLS x 2. A character cannot be scaled, so the only")
	fmt.Println("    way up is a bitmap pass -- and a bitmap CAN hold the glyph vocabulary.")
	fmt.Println("    Drawn with PlotOn and the coat as ground, so a hole is a socket and not")
	fmt.Println("    the sea (his 12:49 report, and what EyeFillSocket already exists for).")
	fmt.Println()
	names := []string{"alert 'O'", "open 'o'", "shut '-'", "done '^'"}
	bms := [][]string{eyeAlertBM, eyeOpenBM, eyeShutBM, eyeDoneBM}
	var cols [][]string
	for _, b := range bms {
		cols = append(cols, render(b))
	}
	fmt.Println(sideBySide(cols, names, 4, false))
	fmt.Println("    ...and each one stamped on the 16x9 body, mirrored:")
	fmt.Println()
	body := renderM(join(pickUpperRows(), pickLower))
	var c2 [][]string
	for _, b := range bms {
		q := render(b)
		c2 = append(c2, stamp(stamp(body, q, 5, 1), q, 9, 1))
	}
	fmt.Println(sideBySide(c2, names, 3, false))
	fmt.Println("    eye ink as a fraction of the face box:")
	for _, r := range []struct {
		lab  string
		w, h int
	}{{"shipped 12x7, glyph", 12, 7}, {"A 14x9, glyph", 14, 9}, {"A 14x9, 2x2 bitmap", 14, 9},
		{"PICK 16x9, glyph", 16, 9}, {"PICK 16x9, 2x2 bitmap", 16, 9},
		{"B 16x11, 2x2 bitmap", 16, 11}, {"D 24x14, glyph", 24, 14}, {"D 24x14, 2x2 bitmap", 24, 14}} {
		n := 1
		if strings.Contains(r.lab, "bitmap") {
			n = 4
		}
		fmt.Printf("        %-24s %d of %3d cells = %.2f%%\n", r.lab, n, r.w*r.h, 100*float64(n)/float64(r.w*r.h))
	}
}
