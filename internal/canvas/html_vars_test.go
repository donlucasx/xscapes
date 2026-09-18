package canvas

import (
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// A palette colour can be written through a CSS variable, with the real
// colour as the fallback, so a page with two themes can re-ink one colour
// per theme: the companions' eye glyphs sit in holes that show the page's
// ground, and their mint vanishes on a light one (his note, 2026-09-17).
func TestAPaletteColourCanCarryAThemeVariable(t *testing.T) {
	mint := term.RGB{R: 168, G: 236, B: 176}
	pal := &HTMLPalette{Vars: map[term.RGB]string{mint: "eye"}}
	pal.class(mint, term.RGB{})
	pal.class(term.RGB{R: 10, G: 20, B: 30}, term.RGB{})
	css := pal.CSS()
	if strings.Count(css, "color:var(--eye,#a8ecb0)") != 1 {
		t.Fatalf("the mint did not go through its variable:\n%s", css)
	}
	if !strings.Contains(css, "color:#0a141e") {
		t.Fatalf("an ordinary colour lost its plain hex:\n%s", css)
	}
}
