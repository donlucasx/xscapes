// Command c-lean draws DIRECTION C of the "come closer before the ask" study:
// no scale change at all. The crab stays in its 12x7 box and the approach is
// carried entirely by the upper half -- stalks, claws, and where the eye sits.
//
//	go run ./notes/s28-closer/c-lean
package main

import (
	"fmt"
	"strings"
)

func main() {
	check()

	fmt.Println("=== BASELINE: what ships today ===")
	shipped := []pose{
		{name: "rest", upper: crabRest, eyes: [2]int{4, 7}, row: 2, glyph: '-'},
		{name: "work", upper: crabWork, eyes: [2]int{4, 7}, row: 2, glyph: 'o'},
		{name: "ask", upper: crabAsk, eyes: [2]int{4, 7}, row: 2, glyph: 'O'},
	}
	strip("shipped states", shipped, true, nil)

	fmt.Println("\n=== THE APPROACH, five beats (mirrored, as shipped) ===")
	strip("approach", approach(), true, nil)
	metrics(approach())

	fmt.Println("\n=== the same five, authored (unmirrored) ===")
	strip("approach", approach(), false, nil)

	fmt.Println("\n=== A/B: shipped ask vs lean ask, same place, same size ===")
	pair(pose{name: "shipped ask", upper: crabAsk, eyes: [2]int{4, 7}, row: 2, glyph: 'O'},
		pose{name: "lean ask", upper: open, eyes: [2]int{3, 8}, row: 1, glyph: 'O'}, true)

	fmt.Println("\n=== REJECTED 1: claws round in front of the shell ===")
	pair(pose{name: "work", upper: crabWork, eyes: [2]int{4, 7}, row: 2, glyph: 'o'},
		pose{name: "front claws", upper: frontClaw, lower: frontClawLower,
			eyes: [2]int{4, 7}, row: 2, glyph: 'O'}, true)

	fmt.Println("\n=== RIVAL ASKS: the box makes claws and eyes fight for the columns ===")
	strip("rival asks", []pose{
		{name: "work", upper: crabWork, eyes: [2]int{4, 7}, row: 2, glyph: 'o'},
		{name: "lean", upper: open, eyes: [2]int{3, 8}, row: 1, glyph: 'O'},
		{name: "bigclaw", upper: bigClaw, eyes: [2]int{4, 7}, row: 1, glyph: 'O'},
		{name: "bulk", upper: bulk, eyes: [2]int{4, 7}, row: 1, glyph: 'O'},
	}, true, nil)
	area([]pose{
		{name: "rest", upper: crabRest, eyes: [2]int{4, 7}, row: 2, glyph: '-'},
		{name: "work", upper: crabWork, eyes: [2]int{4, 7}, row: 2, glyph: 'o'},
		{name: "shipped ask", upper: crabAsk, eyes: [2]int{4, 7}, row: 2, glyph: 'O'},
		{name: "worried", upper: crabWorried, eyes: [2]int{4, 7}, row: 2, glyph: 'o'},
		{name: "lean ask", upper: open, eyes: [2]int{3, 8}, row: 1, glyph: 'O'},
		{name: "arm forward", upper: leanAsk, eyes: [2]int{3, 8}, row: 1, glyph: 'O'},
		{name: "bigclaw ask", upper: bigClaw, eyes: [2]int{4, 7}, row: 1, glyph: 'O'},
		{name: "bulk ask", upper: bulk, eyes: [2]int{4, 7}, row: 1, glyph: 'O'},
		{name: "max fill", upper: maxFill, eyes: [2]int{4, 7}, row: 2, glyph: 'O'},
	})

	fmt.Println("\n=== THE CEILING: every fillable cell filled ===")
	pair(pose{name: "work", upper: crabWork, eyes: [2]int{4, 7}, row: 2, glyph: 'o'},
		pose{name: "max fill", upper: maxFill, eyes: [2]int{4, 7}, row: 2, glyph: 'O'}, true)

	ap := approach()
	fmt.Println("\n=== IN THE SCENE ===")
	sideBySide("left: work, home column        |   right: lean ask, SAME column, SAME row",
		beach(ap[0], 34, 14, 5, 20, 0, ""),
		beach(ap[4], 34, 14, 5, 20, 0, "needs your input"))
	sideBySide("left: work, home column        |   right: lean ask, paced 6 in, dropped 2 rows",
		beach(ap[0], 34, 14, 5, 20, 0, ""),
		beach(ap[4], 34, 14, 5, 14, 2, "needs your input"))
	sideBySide("left: work, home column        |   right: lean ask, dropped 2 rows ONLY",
		beach(ap[0], 34, 14, 5, 20, 0, ""),
		beach(ap[4], 34, 14, 5, 20, 2, "needs your input"))

	sideBySide("left: lean ask (open pincer, wide eyes)   |   right: bigclaw ask",
		beach(ap[4], 34, 14, 5, 14, 2, "needs your input"),
		beach(pose{name: "bigclaw", upper: bigClaw, eyes: [2]int{4, 7}, row: 1, glyph: 'O'},
			34, 14, 5, 14, 2, "needs your input"))

	fmt.Println("\n=== THE PACE SNAP, which is already shipped ===")
	paceSnap()

	fmt.Println("\n=== THE APPROACH IN THE SCENE, beat by beat ===")
	filmstrip(ap, 26, 12, 4, 12, []int{0, -1, -3, -5, -6}, []int{0, 0, 1, 1, 2}, "needs you")

	fmt.Println("\n=== 40x12, the narrow floor (CatX = 40 - margin 3 - 12 = 25) ===")
	sideBySide("left: work, home   |   right: lean ask, paced 2 in, dropped 2",
		beach(ap[0], 40, 12, 4, 25, 0, ""),
		beach(ap[4], 40, 12, 4, 23, 2, "needs input"))

	fmt.Println("\n=== LAST BEAT: is the forward arm worth its noise? ===")
	pair(pose{name: "open + O only", upper: open, eyes: [2]int{3, 8}, row: 1, glyph: 'O'},
		pose{name: "arm forward", upper: leanAsk, eyes: [2]int{3, 8}, row: 1, glyph: 'O'}, true)

	fmt.Println("\n=== THE FINAL POSE, printed as source rows ===")
	fmt.Println("upper (24x12):")
	fmt.Print(literal(open))
	fmt.Println("lower (24x16): unchanged, crabLower")
	fmt.Println("\nthe pose as the scape paints it (mirrored, eyes plotted, no lift):")
	for _, r := range render(pose{upper: open, eyes: [2]int{3, 8}, row: 1, glyph: 'O'}, true, 0) {
		fmt.Println(r)
	}
}

// check proves the builder before any of its output is trusted: mirror12 of the
// left half must reproduce the shipped crabWork character for character.
func check() {
	for i := range crabWork {
		if genWork[i] != crabWork[i] {
			panic(fmt.Sprintf("mirror12 does not reproduce crabWork at row %d:\n got %q\nwant %q",
				i, genWork[i], crabWork[i]))
		}
	}
	for _, rows := range [][]string{halt, wide, open, leanAsk, frontClaw, maxFill} {
		for _, r := range rows {
			if len([]rune(r)) != 24 {
				panic("ragged draft row: " + r)
			}
		}
	}
}

// approach is the five beats, each changing one thing: the eye climbs a cell,
// the claw opens, the eyes step apart, the glyph grows.
func approach() []pose {
	return []pose{
		{name: "work", upper: crabWork, eyes: [2]int{4, 7}, row: 2, glyph: 'o'},
		{name: "halt", upper: halt, eyes: [2]int{4, 7}, row: 1, glyph: 'o'},
		{name: "open", upper: openNarrow, eyes: [2]int{4, 7}, row: 1, glyph: 'o'},
		{name: "wide", upper: open, eyes: [2]int{3, 8}, row: 1, glyph: 'o'},
		{name: "ask", upper: open, eyes: [2]int{3, 8}, row: 1, glyph: 'O'},
	}
}

// metrics is the honest accounting: how much of the painted picture actually
// moves between the first beat and the last, and where.
func metrics(ps []pose) {
	base := render(ps[0], true, 0)
	fmt.Printf("\n%-8s %6s %8s %8s %8s  %s\n", "beat", "cells", "top3", "shell4", "changed", "changed cells vs beat 0")
	for _, p := range ps {
		r := render(p, true, 0)
		top3, shell := 0, 0
		for i := 0; i < 3; i++ {
			top3 += inkCells([]string{r[i]})
		}
		for i := 3; i < 7; i++ {
			shell += inkCells([]string{r[i]})
		}
		diff := 0
		for i := 0; i < 7; i++ {
			a, b := []rune(base[i]), []rune(r[i])
			for j := 0; j < 12; j++ {
				x, y := ' ', ' '
				if j < len(a) {
					x = a[j]
				}
				if j < len(b) {
					y = b[j]
				}
				if x != y {
					diff++
				}
			}
		}
		fmt.Printf("%-8s %6d %8d %8d %8d  %s\n", p.name, inkCells(r), top3, shell, diff,
			strings.Repeat("#", diff))
	}
	fmt.Println("top3 = the three cell rows the upper half owns; shell4 = the four it cannot touch.")
}

// area is the apparent-size table: quarter cells painted, which is what the
// question "does it look closer" actually turns on. 336 is a solid 12x7 box.
func area(ps []pose) {
	base := quads(ps[1])
	fmt.Printf("\n%-13s %7s %7s   %s\n", "pose", "quads", "vs work", "")
	for _, p := range ps {
		q := quads(p)
		fmt.Printf("%-13s %7d %6.1f%%   %s\n", p.name, q, 100*float64(q-base)/float64(base),
			strings.Repeat("#", q/6))
	}
	fmt.Println("a 16x9 box -- the room the measured envelope allows -- would be 576 quads, +109%.")
}
