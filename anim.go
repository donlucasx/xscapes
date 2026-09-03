package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

// animPage renders each state as a real frame sequence and cycles it in the
// browser. A still cannot show breathing, and breathing is the whole point.
func animPage(seed int64, frames int, fps float64) string {
	cat := companion.NewCat()
	cw, ch := cat.Size()

	states := []struct {
		st   companion.State
		note string
		say  string
	}{
		{companion.Resting, "slow breath, tail barely moves, eyes half shut", ""},
		{companion.Working, "faster breath, tail sweeping, eyes open, occasional blink", ""},
		{companion.NeedsYou, "quick breath, tail flicking, wide eyes, ask balloon", "allow Bash?"},
		{companion.Done, "slow breath, tail held high and still, content eyes, soft knock", "tests passed"},
	}

	var b strings.Builder
	b.WriteString(`<style>` +
		`.stage{position:relative}.fr{display:none}.fr:first-child{display:block}` +
		`.row{display:flex;gap:26px;align-items:center}` +
		`</style>`)
	b.WriteString(`<h1>xscapes &mdash; companion, animated</h1>`)

	for _, s := range states {
		var st strings.Builder
		for i := 0; i < frames; i++ {
			t := float64(i) / fps
			c := canvas.New(72, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			act := scape.Activity{}
			if s.st != companion.Resting {
				act = scape.Activity{Working: true, Level: 0.65}
			}
			scape.NewShore(seed, false).Update(c, t, act)

			top := c.H - 2 - ch
			cat.Draw(c.Near(), 6, top, t, s.st)
			if s.say != "" {
				rows, col := companion.Bubble(s.say), bubbleAskCol
				if s.st == companion.Done {
					rows, col = companion.DoneBubble(s.say), bubbleCol
				}
				(&companion.Sprite{Rows: rows, Body: col, Opaque: true}).
					Draw(c.Near(), 6+cw-2, top-len(rows))
			}
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragment(13))
		}
		fmt.Fprintf(&b,
			`<div class="card"><div class="meta"><div class="nm">%s</div>`+
				`<div class="rg">%d frames &middot; %.0f fps</div><div class="nt">%s</div></div>`+
				`<div class="stage">%s</div></div>`,
			s.st, frames, fps, s.note, st.String())
	}

	// Walking: there and back, so the loop is seamless and both facings show.
	{
		ww, wh := cat.WalkSize()
		const steps = 48
		x0, x1 := 4, 72-ww-4
		var st strings.Builder
		for i := 0; i < steps; i++ {
			half := steps / 2
			var px int
			var dir int
			if i < half {
				px, dir = x0+(x1-x0)*i/half, 1
			} else {
				px, dir = x1-(x1-x0)*(i-half)/half, -1
			}
			c := canvas.New(72, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			scape.NewShore(seed, false).Update(c, float64(i)/fps,
				scape.Activity{Working: true, Level: 0.5})
			// Phase tracks distance walked, not the clock, so the feet do not skate.
			cat.DrawWalk(c.Near(), px, c.H-2-wh, float64(px)*0.8*float64(dir), dir)
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragment(13))
		}
		fmt.Fprintf(&b,
			`<div class="card"><div class="meta"><div class="nm">walking</div>`+
				`<div class="rg">%d frames &middot; %.0f fps</div>`+
				`<div class="nt">side view, four legs on diagonal pairs, body bobs at mid-stride; `+
				`turns around at each end</div></div>`+
				`<div class="stage">%s</div></div>`, steps, fps, st.String())
	}

	fmt.Fprintf(&b, `<script>
document.querySelectorAll('.stage').forEach(function(st){
  var fr = Array.prototype.slice.call(st.children), i = 0;
  setInterval(function(){
    fr[i].style.display = 'none';
    i = (i + 1) %% fr.length;
    fr[i].style.display = 'block';
  }, %d);
});
</script>`, int(1000/fps))
	return canvas.HTMLPage("xscapes — companion animated", b.String())
}
