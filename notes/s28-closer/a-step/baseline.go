package main

import "fmt"

// The shipped art, copied here so this draft can be rendered and diffed
// without touching internal/companion (other agents are in this tree).
// Verbatim from internal/companion/crab.go, 2026-09-10.

var farLower = []string{
	"########################",
	"########################",
	".######################.",
	"..####################..",
	"...##################...",
	"..##..############..##..",
	"..##..############..##..",
	".##....##########....##.",
	".##....##########....##.",
	"##......########......##",
	"##......########......##",
	"##.......######.......##",
	".#.......######.......#.",
	".#........####........#.",
	".#....................#.",
	".#....................#.",
}

var farAsk = []string{
	"..................##..##",
	"..................##..##",
	"..................######",
	"..................######",
	"##..##............######",
	"##..##.............####.",
	"######..............##..",
	".####...............##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

func baseline() {
	rule("THE SHIPPED CRAB, so every candidate is read against it")
	checkWidths("farAsk", farAsk)
	checkWidths("farLower", farLower)
	show("shipped ask pose", join(farAsk, farLower))
	showMirrored("shipped ask pose", join(farAsk, farLower))
	fmt.Println("eyes: single-cell glyphs plotted on top at cells {4,7}, row 2.")
	fmt.Println("widest ink run per source row (the shell's taper):")
	fmt.Println(" ", rowRuns(join(farAsk, farLower)))
}

func nnExperiment() {
	rule("REJECTED FIRST: nearest-neighbour scaling 24x28 -> 28x36")
	up := nearest(join(farAsk, farLower), 28, 36)
	checkWidths("nn", up)
	show("nearest-neighbour 14x9", up)

	badSrc, _ := symmetry(join(farAsk, farLower))
	badNN, firstNN := symmetry(up)
	fmt.Printf("rows whose ink is NOT left-right symmetric:\n")
	fmt.Printf("  hand-authored 24x28 : %d of 28   (the ask pose is asymmetric BY DESIGN -- one claw up)\n", badSrc)
	fmt.Printf("  nearest-neighbour   : %d of 36   first at row %d\n", badNN, firstNN)

	fmt.Printf("\nsame test on a pose that IS symmetric in the source -- the shell and legs:\n")
	badL, _ := symmetry(farLower)
	nnL := nearest(farLower, 28, 20)
	badNL, firstNL := symmetry(nnL)
	fmt.Printf("  farLower  24x16 : %d of 16 asymmetric rows\n", badL)
	fmt.Printf("  scaled    28x20 : %d of 20 asymmetric rows, first at row %d\n", badNL, firstNL)
	show("nearest-neighbour lower half", nnL)
	fmt.Println("widest run per row, source vs scaled:")
	fmt.Println("  src:", rowRuns(farLower))
	fmt.Println("  nn :", rowRuns(nnL))
}
