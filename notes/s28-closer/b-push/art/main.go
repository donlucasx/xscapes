// Command art renders direction B's push-in art through the REAL
// companion.ParseBitmap / ToQuadrant, because art that was not rendered is a
// guess. Every silhouette printed here came out of the shipped medium.
//
//	go run ./notes/s28-closer/b-push/art
package main

import "fmt"

func main() {
	shipCheck()
	todayRef()

	fmt.Println("=== THE LADDER: today 12x7 -> mid 14x9 -> big 16x11 ===")
	fmt.Println()
	q7 := show("today  ask (12x7)", join(shipAsk, shipLower))
	q9 := show("mid    ask (14x9)", join(midAsk, midLower))
	q11 := show("big    ask (16x11)", join(bigAsk, bigLower))

	fmt.Println("=== THE EYES, as their own Sprite pass ===")
	fmt.Println("today: a single glyph 'O'. mid: 2 cells x 1. big: 2 cells x 2.")
	fmt.Println()
	showEye("mid eye  (4x4 px)", midEye)
	showEye("big eye  solid (4x8 px)", bigEyeSolid)
	showEye("big eye  ring  (4x8 px)", bigEyeRing)

	fmt.Println("=== THE POSE AS IT ACTUALLY DRAWS, body pass then eye pass ===")
	fmt.Println("(the eye is printed as E so the two passes can be told apart)")
	fmt.Println()
	frame("today, eye = one glyph at cells {4,7} row 2",
		overlayGlyph(q7, [2]int{4, 7}, 2, 'O'))
	frame("mid, eye = 2x1 bitmap at cells {4,8} row 2",
		overlay(q9, quad(midEye), midEyeCells, midEyeRow))
	frame("big, eye = 2x2 SOLID bitmap at cells {5,9} rows 2-3",
		overlay(q11, quad(bigEyeSolid), bigEyeCells, bigEyeRow))
	frame("big, eye = 2x2 RING (the hole shows the sea, as the shipped eyes do)",
		overlay(q11, quad(bigEyeRing), bigEyeCells, bigEyeRow))

	fmt.Println("=== MIRRORED, which is the shipped composition ===")
	fmt.Println("the raised claw has to end up on the LEFT, pointing at the transcript.")
	fmt.Println()
	frame("big ask, mirrored, with the eye pass mirrored too",
		overlay(mirrored(join(bigAsk, bigLower)), quad(bigEyeSolid),
			mirrorCells(bigEyeCells, 16, 2), bigEyeRow))
	frame("today ask, mirrored", overlayGlyph(mirrored(join(shipAsk, shipLower)),
		mirrorCells([2]int{4, 7}, 12, 1), 2, 'O'))

	fmt.Println("=== SIDE BY SIDE, feet aligned, which is how he would see the change ===")
	fmt.Println()
	sideBySide(
		[]string{"today 12x7", "mid 14x9", "big 16x11"},
		[][]string{
			overlayGlyph(mirrored(join(shipAsk, shipLower)), mirrorCells([2]int{4, 7}, 12, 1), 2, 'O'),
			overlay(mirrored(join(midAsk, midLower)), quad(midEye), mirrorCells(midEyeCells, 14, 2), midEyeRow),
			overlay(mirrored(join(bigAsk, bigLower)), quad(bigEyeSolid), mirrorCells(bigEyeCells, 16, 2), bigEyeRow),
		})

	fmt.Println("=== THE APPROACH, frame by frame ===")
	fmt.Println("A size is a whole hand-drawn sprite, so the push-in is three rungs, not")
	fmt.Println("a scale. Between rungs the only motion the medium has is the half-cell")
	fmt.Println("lift the breathing already uses -- two source rows.")
	fmt.Println()
	approach()

	fmt.Println("=== WORKING POSE at the big size, which is what it grows out of ===")
	fmt.Println()
	frame("big work (16x11)", overlay(quad(join(bigWork, bigLower)), quad(bigEyeSolid),
		bigEyeCells, bigEyeRow))
}

func join(a, b []string) []string { return append(append([]string{}, a...), b...) }
