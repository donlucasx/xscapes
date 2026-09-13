package main

// Candidate near eyes for the crab, drafted 2026-09-11 from his screenshot of
// his own live ask: "can we mockup an alternative approach for the crab close
// up where the eyes read softer? maybe thiner lines or an alt approach where
// the companion reads gentler yet ready". GENERATED from the drafting round.
//
// Each is 4 source columns by 8 source rows -- SIXTEEN SUBPIXELS after the
// halving -- which is the whole budget this decision has.

import "github.com/donlucasx/xscapes/internal/term"

type crabEyeAlt struct {
	Key, Name, Idea string
	Alert           []string
	Col             term.RGB
	ColWhy          string
	Still           string
	Risks           []string
}

var crabEyeAlts = []crabEyeAlt{
	{
		Key:  "round-off",
		Name: "Round Off",
		Idea: "The shipped ring with its four corners dropped: at 2x2 cells every corner cell turns from a three-quarter block into a diagonal half, so the coat wraps into all four corners and the eye reads as a round O set *into* the face instead of a bright square stuck *on* it.",
		Alert: []string{
			".##.",
			".##.",
			"#..#",
			"#..#",
			"#..#",
			"#..#",
			".##.",
			".##.",
		},
		Col:    term.RGB{R: 232, G: 252, B: 226},
		ColWhy: "Unchanged, and I can now show that changing it is a trap rather than a taste call. EyeAlert (232,252,226) has chroma 26; neutralChroma is 30. It sits FOUR POINTS under the cliff, so Saturate leaves it alone and it quantises to cube 194 = (215,255,215), a pale mint. The working bead (168,236,176) has chroma 68, takes the 2.6 GlyphBoost, and lands on 121 = (135,255,175). Now the counterintuitive part, measured: any \"softer, less saturated mint\" crosses chroma 30 and therefore G",
		Still:  "Rendered through the real ToQuadrant, the O is `▞▚` over `▚▞`; the bead is `▗▖` over `▝▘`; the lid is two spaces over `▀▀`. VS THE WORKING BEAD: the bead is four single quarter-blocks meeting at the cell junction -- one small dot at the centre, 1/4 ink per cell, confined to the middle 2x2 of the footprint. The O is four diagonal halves forming a closed outline across the FULL 4x4, 2/4 ink per cell. Outline versus dot, and all eight cells differ in glyph -- there is no frame o",
		Risks: []string{
			"THE ARGUMENT, from the medium: at 4x4 there is no such thing as a gentle corner. The corner subpixel is the same size as the wall, so the turn happens in one step at 90 degrees -- that IS a rectangle. You cannot bevel at this resolution, you can only omit. Four of the ring's twel",
			"WHAT THE FOUR WALL SEGMENTS DO AT CELL LEVEL: shipped, every cell is a three-quarter block (▛▜/▙▟) and four of those butted together read as one solid square with a bite out of it -- the outer boundary is a perfect axis-aligned square because every cell's outer corner is full. Th",
			"IT MAY SIMPLY NOT BE ENOUGH. This is four subpixels. That is the bet: a change this small cannot break anything, so if it is enough it wins on restraint. If he looks and still says hard, the next lever is the hole or a lid, not the corners -- and nothing here has to be undone fir",
		},
	},
	{
		Key:  "sea-glass",
		Name: "Sea Glass",
		Idea: "He sees the same wide-open ring in the same place, but painted in tumbled sea glass instead of near-white -- it stops reading as a lit rectangle stuck on the shell and starts reading as a green eye, because dropping 40 points of luminance against the salmon is what makes a one-subpixel stroke read as a thin line rather than a hard edge.",
		Alert: []string{
			"####",
			"####",
			"#..#",
			"#..#",
			"#..#",
			"#..#",
			"####",
			"####",
		},
		Col:    term.RGB{R: 184, G: 210, B: 184},
		ColWhy: "Checked against the real quantiser rather than by eye in RGB. The eye is a GLYPH, so it goes Saturate(GlyphBoost=2.6) then Index256; the coat is a BACKGROUND and goes Index256Keeping. Shipped EyeAlert 232,252,226 has chroma 26, which is UNDER neutralChroma=30, so the boost never touches it and it quantises to cube 194 = 215,255,215: a near-white at L 238.6 against a coat painted at L 171.0, i.e. +67.6 -- as bright as the moon, which is the shout. My 184,210,184 also has chrom",
		Still:  "The shape is untouched, so every still-frame guarantee the shipped ring bought is intact, and I rendered all three states side by side at true terminal pixel size (subpixel 7x15, so the eye is a 28x60 upright block) to look rather than assert. Against the working bead: the ring lights 12 of 16 subpixels and reaches all four corners of its 2x2 cells (glyphs Y/Z pattern, upper row two 7/8-full quadrant blocks), while the bead lights 4 centred subpixels and never touches a cell ",
		Risks: []string{
			"EyeAlert is shared, so editing the var also dims the 1x eye at cat.go:305, crab.go:363 and catalts.go:344 -- and at 1x the eye is ONE cell with no shape channel, so the near-white is doing the whole job there; the safe form is a near-only colour read by nearEye, but that breaks t",
			"The ask is now DIMMER and much less saturated than the working bead (L 198.6 / chroma 40 against L 210.2 / chroma 120), so the salience hierarchy rests entirely on the ring's 3x ink area and full-cell footprint; if that reads wrong to him the whole direction fails and the answer ",
			"On the retreat (the crab leaves NeedsYou and walks back) the eye now BRIGHTENS as the ask is answered -- sage ring to vivid bead -- which I could not test in a still and which may read as the animal perking up, or as the ask having been the quiet state.",
		},
	},
	{
		Key:  "the-high-pupil",
		Name: "The high pupil",
		Idea: "Both eyes keep the same full tile, but the coat showing through stops being a tall slot down the middle and becomes one small square sitting high inside the eye, so the ask reads as a pupil looking up at you instead of a bright rectangle with a bar through it.",
		Alert: []string{
			"####",
			"####",
			"#..#",
			"#..#",
			"####",
			"####",
			"####",
			"####",
		},
		Col:    term.RGB{R: 232, G: 252, B: 226},
		ColWhy: "Measured, not guessed: I ran the real path (Saturate(GlyphBoost=2.6) then Index256) over a sweep of mint-ish inks. 232,252,226 has chroma 26, under neutralChroma 30, so the boost leaves it alone and it lands on cube (215,255,215). There is no softer mint beside it. Dim it at the same hue and it falls into the grey ramp: 206,232,206 -> grey 218, 200,228,200 -> grey 218. Its bright neighbours are white (255,255,255), pale cyan (215,255,255), cream (255,255,215), none gentler. T",
		Still:  "Frozen, it is the only eye whose mint reaches all four corners of its 2x2 cell footprint: it packs to the glyphs ▛▜ over ██, a full tile with one square notch high on the inner top edge. The working bead packs to ▗▖ over ▝▘, a single small block floating with every outer corner salmon; nearEyeShut packs to two blank cells over ▀▀, a thin bar with the whole upper cell row coat. Lit subpixels are 14, 4 and 4, and the two fours differ in placement (centred square vs full-width b",
		Risks: []string{
			"It is BRIGHTER than what shipped, and that is the honest cost of this direction. The ring already carries the largest hole a closed 4x4 outline allows, so every inside-only edit lights MORE subpixels: 14 of 16 here against 12. If his 'hard' is about raw luminous area, only the ou",
			"The gentleness rests entirely on the pupil reading as round. At his cell geometry (14 x 30 px) a subpixel is 7 x 15, so the shipped 2x2 hole is a 14 x 30 tall SLOT while this 2x1 hole is 14 x 15, near square - the roundest dot the medium can make. If he does not read that square ",
			"Cartoon grammar cuts against it: a DILATED pupil reads friendly, a small one reads startled. This shrinks the pupil. I am betting 'a pupil at all' beats 'a big slot', which the renders support, but it is a bet.",
		},
	},
	{
		Key:  "the-hood",
		Name: "The hood",
		Idea: "A thin near-white brow laid across the top of the eye with two hairline legs hanging from its ends and no floor underneath, so the salmon stalk runs unbroken up through the middle and out the open bottom — the eye stops being a lit rectangle with a bar trapped in it and becomes an outline the animal shows through, while the raised brow keeps it plainly attending.",
		Alert: []string{
			"####",
			"####",
			"#..#",
			"#..#",
			"#..#",
			"#..#",
			"....",
			"....",
		},
		Col:    term.RGB{R: 232, G: 252, B: 226},
		ColWhy: "Unchanged, and I measured why rather than guessing. DetectProfile returns Profile256 on EVERY terminal unless XSCAPES_COLOR is set (color.go, ruled 2026-09-04), so the cube is the only path that ships. Through Profile256.Quantise(fg=true): shipped 232,252,226 -> index 194 -> painted 215,255,215; working eyeCol 168,236,176 -> index 121 -> painted 135,255,175. Every softer mint I probed (206,240,206 / 196,232,200 / 214,244,212) collapses onto ONE neighbour, index 157 -> painted",
		Still:  "Rendered through ToQuadrant the hood is `▛▜` over `▘▝`: ink hard against the TOP edge of the eye tile across its full width, two 1-subpixel legs (7px) hanging from its ends down 30px, and the bottom 15px of the tile entirely coat. Against the working bead (`▗▖` / `▝▘`) there is no contest — the bead touches no edge of the tile at all, it is a 14x30 island of green floating in salmon, while the hood is near-white and is the only alert art that reaches the tile's top edge; the ",
		Risks: []string{
			"Lit area drops from 12 subpixels to 8, a third less mint, and the ask has to pull his eye across a terminal. In the full-body render it still reads (it is the only near-white thing on the animal) but I judged that on a still, not in his peripheral vision during real work. If it r",
			"THE EYE BITMAP IS NOT MIRRORED. drawCrabNear mirrors the eye's CELL POSITION (ex = bx + w - 2 - e) and plots the art as authored, so an asymmetric eye points the same way on both facings. The hood is left-right symmetric and is safe, but this is what killed the crescent and every",
			"The blink still fires during NeedsYou (nearEye swaps art to nearEyeShut but keeps EyeAlert), so the ask now blinks from a bar at the TOP to a bar at the MIDDLE -- the lid reads as dropping rather than the ring collapsing. I only looked at stills; that transition wants one look in",
		},
	},
	{
		Key:  "the-hooded-ring",
		Name: "the hooded ring",
		Idea: "A person sees two eyes with a heavy upper lid, a salmon aperture under it and a lower-lid line, sitting BELOW the stalk's own salmon tip instead of capping it — a third less ink than the shipped ring (8 subpixels of 16 against 12), no square corners, and a stalk that still ends in coat, so it reads as an eye half-open and looking at you rather than two lit rectangles stuck on the shell.",
		Alert: []string{
			"....",
			"....",
			"####",
			"####",
			"#..#",
			"#..#",
			".##.",
			".##.",
		},
		Col:    term.RGB{R: 186, G: 238, B: 196},
		ColWhy: "Measured through the repo's own quantiser (internal/term/color.go, copied out and run, not inferred): eyeCol=157, EyeAlert=194, coat=210. A \"slightly softer white\" is a SILENT NO-OP on a 256-colour terminal — 214,246,214 / 205,238,222 / 198,240,206 all quantise to 194, identical to the shipped colour. The cube gives exactly one usable step down: 186,238,196 → 158 = (175,255,215), red 40 lower than the shipped (215,255,215) and still 40 of blue clear of working's (175,255,175)",
		Still:  "Against the working bead: the bead is ONE mark, one cell wide, centred on the stalk (14x30 px) in green — cells ▗▖/▝▘. The hooded ring is THREE tiers spanning the stalk's full 28px in near-white: a lid bar, an enclosed salmon aperture, a chin — cells ▄▄/▚▞. No glyph in common. Against nearEyeShut: shut is a single 28x15 bar in the LOWER cell row with coat above and below, and its upper cell row is BLANK (\" \"). The hooded ring's bar is one subpixel row higher, in the UPPER cel",
		Risks: []string{
			"SQUINT/SLEEPY is the real risk and here is exactly where the line is. The aperture is one subpixel row (14x15 px). Fill it — 0000/1111/1111/0110 — and you have a bar over a chin, an '=' sign, which is nearEyeShut with a second line; go one further, 0000/0000/1111/1111, and it is ",
			"A lid is the most expensive stroke this medium has. Source rows are OR'd in pairs before packing, so the thinnest horizontal stroke is a full subpixel row = 15px, while the thinnest vertical stroke is one source column = 7px. A lid is therefore always 2.1x thicker than the eye's ",
			"Subpixel ROW 0 of the eye tile is the STALK'S SILHOUETTE, not a place for a brow. This killed my best-looking candidate: a full-width brow at row 0 over a bead at rows 2-3 looked great on a flat coat swatch and, rendered on the actual rung-2 crab (crabAsk+crabLower doubled, eyeCe",
		},
	},
}
