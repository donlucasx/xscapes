package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// meowCat is ref 4 transcribed character-exact from the source image, rather
// than paraphrased from memory of it -- which is what went wrong the first time.
// Only the ambiguous-width overbar and degree sign were substituted.
var meowCat = companion.Sprite{
	Name: "cat, meow", Register: companion.Line, Source: "ref 4, transcribed",
	Note: "hand-drawn, faithful this time",
	Rows: []string{
		`   |\__/|`,
		`  (_ ^-^)`,
		`    )   (`,
		` _  )   (`,
		`((( /     \`,
		` (  )  || ||`,
		` '----' '--''--'`,
	},
	Body: term.RGB{R: 232, G: 224, B: 206},
}

func compareSheet(seed int64) string {
	bm := companion.ParseBitmap(companion.PixelCat)
	warm := term.RGB{R: 232, G: 224, B: 206}

	type variant struct {
		name, note string
		rows       []string
	}
	variants := []variant{
		{"one char per cell", "84 cells &middot; 1:2 stretched pixels &middot; all ASCII", bm.ToChars()},
		{"quadrants", "168 px &middot; square pixels &middot; needs 4 ambiguous-width blocks", bm.ToQuadrant()},
		{"braille", "672 px &middot; square pixels &middot; zero ambiguous runes", bm.ToBraille()},
	}

	var b strings.Builder
	b.WriteString(`<h1>xscapes &mdash; one drawing, three renderings</h1>`)

	// The source, so it is clear all three come from the same 24x28 art.
	src := canvas.New(bm.W+2, bm.H+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for y := 0; y < src.H; y++ {
		for x := 0; x < src.W; x++ {
			src.SetBG(x, y, term.RGB{R: 27, G: 27, B: 33})
		}
	}
	var srcRows []string
	for _, r := range companion.PixelCat {
		srcRows = append(srcRows, strings.ReplaceAll(r, ".", " "))
	}
	(&companion.Sprite{Rows: srcRows, Body: warm}).Draw(src.Near(), 1, 1)
	fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">source</div>`+
		`<div class="rg">24&times;28 bitmap</div><div class="nt">authored once; every rendering below is generated from this</div></div>`+
		`<div class="big"><div class="lbl">pixel art</div>%s</div></div>`, src.HTMLFragment(11))

	for _, v := range variants {
		b.WriteString(card(v.name, v.note, v.rows, warm, seed))
	}
	b.WriteString(card(meowCat.Name+" (hand-drawn)", meowCat.Note, meowCat.Rows, warm, seed))
	return canvas.HTMLPage("xscapes — rendering comparison", b.String())
}

func card(name, note string, rows []string, col term.RGB, seed int64) string {
	s := companion.Sprite{Rows: rows, Body: col}
	sw, sh := s.Size()

	big := canvas.New(sw+2, sh+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for y := 0; y < big.H; y++ {
		for x := 0; x < big.W; x++ {
			big.SetBG(x, y, term.RGB{R: 27, G: 27, B: 33})
		}
	}
	s.Draw(big.Near(), 1, 1)

	live := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	scape.NewShore(seed, false).Update(live, 2.0, scape.Activity{})
	top := live.H - 2 - sh
	s.Draw(live.Near(), 8, top)
	rowsB := companion.Bubble("tests passed")
	(&companion.Sprite{Rows: rowsB, Body: term.RGB{R: 224, G: 228, B: 238}, Opaque: true}).
		Draw(live.Near(), 8+sw-2, top-len(rowsB))

	return fmt.Sprintf(`<div class="card"><div class="meta"><div class="nm">%s</div>`+
		`<div class="rg">%d rows</div><div class="nt">%s</div></div>`+
		`<div class="big"><div class="lbl">enlarged</div>%s</div>`+
		`<div><div class="lbl">on an 80&times;24 shore, 1:1</div>%s</div></div>`,
		name, sh, note, big.HTMLFragment(26), live.HTMLFragment(10))
}
