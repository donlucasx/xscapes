package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

var levels = []int{0, 95, 135, 175, 215, 255}

func luma(c term.RGB) float64 {
	return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
}

func chroma(c term.RGB) int {
	hi, lo := int(c.R), int(c.R)
	for _, v := range []int{int(c.G), int(c.B)} {
		if v > hi {
			hi = v
		}
		if v < lo {
			lo = v
		}
	}
	return hi - lo
}

func run() {
	fmt.Println("THE CUBE'S WARM FLOOR -- entries keeping R > G >= B, by luma:")
	var warm []term.RGB
	for _, r := range levels {
		for _, g := range levels {
			for _, b := range levels {
				if r > g && g >= b {
					warm = append(warm, term.RGB{R: uint8(r), G: uint8(g), B: uint8(b)})
				}
			}
		}
	}
	// The pure reds (255/0/0 and friends) keep the ordering and are useless as
	// sand, so the floor is measured over entries whose green is not zero.
	lowest, at := 255.0, term.RGB{}
	for _, c := range warm {
		if c.G == 0 {
			continue
		}
		if l := luma(c); l < lowest {
			lowest, at = l, c
		}
	}
	fmt.Printf("  darkest usable warm entry is %d/%d/%d at luma %.0f.\n", at.R, at.G, at.B, lowest)
	fmt.Printf("  below it the cube offers only the pure reds, so any darker warm colour\n")
	fmt.Printf("  is nearer a grey-ramp entry than anything with hue.\n\n")

	fmt.Println("hour   dry beach        wet strip        verdict")
	for h := 0; h < 24; h++ {
		p := scape.PaletteAt(float64(h) / 24)
		dry := scape.DryBandExport(p)
		wet := scape.WetBandExport(p)
		verdict := "warm, one rung under the sand"
		if chroma(wet) == 0 {
			verdict = "neutral -- the cube has no warm tone under the sand here"
		}
		if luma(wet) >= luma(dry) {
			verdict += "   <-- NOT DARKER THAN THE SAND"
		}
		fmt.Printf("%02d:00  %3d/%3d/%3d L%3.0f  %3d/%3d/%3d L%3.0f  %s\n",
			h, dry.R, dry.G, dry.B, luma(dry), wet.R, wet.G, wet.B, luma(wet), verdict)
	}

	fmt.Println("\nthe shoreline gradient, sand -> sea, at the hours that matter:")
	for _, h := range []float64{12, 16.42, 19} {
		p := scape.PaletteAt(h / 24)
		fmt.Printf("  %5.2fh  ", h)
		for _, t := range []float64{0, 0.25, 0.5, 0.75, 1} {
			q := term.FromIndex256(term.Lerp(scape.WetBandExport(p), p.SeaNear, t).Index256Keeping())
			mark := " "
			if chroma(q) == 0 {
				mark = "*"
			}
			fmt.Printf("%3d/%3d/%3d%s  ", q.R, q.G, q.B, mark)
		}
		fmt.Println()
	}
	fmt.Println("  (* = neutral grey. At 16:42 three of the five steps are grey.)")
}

// proposed is option 2 from the 2026-09-09 menu, for the table only: hold the
// wet band on a warm cube entry while the DRY beach still draws warm, and let
// it go neutral once the whole beach has, which is the locked rule that
// darkness lives in the backgrounds.
func proposed(sand, wet term.RGB) term.RGB {
	if chroma(term.FromIndex256(sand.Index256Keeping())) == 0 {
		return wet
	}
	if chroma(term.FromIndex256(wet.Index256Keeping())) > 0 {
		return wet
	}
	best, bestD := wet, 1<<62
	for _, r := range levels {
		for _, g := range levels {
			for _, b := range levels {
				if !(r > g && g >= b) || g == 0 {
					continue
				}
				cand := term.RGB{R: uint8(r), G: uint8(g), B: uint8(b)}
				if luma(cand) >= luma(sand) {
					continue
				}
				dr, dg, db := r-int(wet.R), g-int(wet.G), b-int(wet.B)
				if d := dr*dr + dg*dg + db*db; d < bestD {
					bestD, best = d, cand
				}
			}
		}
	}
	return best
}
