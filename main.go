// Command xscapes renders the thinking scene.
//
// Without flags it prints one frame, which is what you want in a pipe. The
// live TUI is a thin wrapper over exactly this renderer.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/envx"
	"github.com/donlucasx/xscapes/internal/notify"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	warnLegacyEnv(os.Args[1:])

	// Subcommands are checked before the flag set is parsed. The renderer has
	// twenty demo flags and the adapters have their own; keeping them in
	// separate flag sets is what stops `emit -tool` from colliding with a
	// future `-t` on the renderer.
	if dispatch(os.Args[1:]) {
		return
	}

	var (
		width   = flag.Int("w", 80, "canvas width")
		height  = flag.Int("h", 24, "canvas height")
		frames  = flag.Int("frames", 1, "how many frames to render")
		plain   = flag.Bool("plain", false, "glyphs only, no colour")
		working = flag.Bool("working", false, "render the working state")
		level   = flag.Float64("level", -1, "activity level 0..1 (default: 0 resting, 0.7 working)")
		seed    = flag.Int64("seed", 7, "scene seed")
		asciiG  = flag.Bool("ascii", false, "ASCII glyphs only, no Unicode")
		fps     = flag.Float64("fps", 20, "time step between rendered frames")
		info    = flag.Bool("info", false, "print the detected colour profile and exit")
		html    = flag.String("html", "", "write the frame to an HTML file instead (for looking at colour)")
		sheet   = flag.String("sheet", "", "write the companion contact sheet to an HTML file")
		compare = flag.String("compare", "", "write the rendering comparison to an HTML file")
		anim    = flag.String("anim", "", "write an animated companion preview to an HTML file")
		strip   = flag.String("strip", "", "write a frame strip (for GIF assembly) to an HTML file")
		layout  = flag.String("layout", "", "write the layout mockups to an HTML file")
		ctxdemo = flag.String("context", "", "write the context-moon demo to an HTML file")
		dayHTML = flag.String("day", "", "write the day cycle, truecolor vs 256, to an HTML file")
		busy    = flag.String("busy", "", "write the activity-level sweep to an HTML file")
		kits    = flag.String("kittens", "", "write the subagent-kitten demo to an HTML file")
		wired   = flag.String("wired", "", "write a simulated session, folded by the real reducer, to an HTML file")
		reel    = flag.String("reel", "", "write a frame strip of one simulated turn (for GIF assembly)")
		colors  = flag.String("colors", "", "write the 256-vs-truecolor study to an HTML file")
		facesHT = flag.String("faces", "", "write the companion face/coat study to an HTML file")
		eyesHT  = flag.String("eyes", "", "write the companion eye-fill study to an HTML file")
		reelAt  = flag.Int("reel-from", 0, "first frame of the reel strip")
		reelN   = flag.Int("reel-count", 40, "how many frames of the reel strip")
		tod     = flag.Float64("tod", 0, "time of day: 0 midnight, .25 dawn, .5 noon, .75 dusk")
		todo    = flag.String("todo", "", "checklist as done/total, e.g. -todo 3/5 (lights the constellation)")
		live    = flag.Bool("live", false, "paint the scape in THIS terminal until Ctrl-C")
		ctxUsed = flag.Float64("ctx", 0, "context used, 0..1 (moon phase and altitude)")
		mode    = flag.String("mode", "working", "strip mode: resting|working|needsyou|walk")
		session = flag.String("session", "", "session to follow in -live (default: $CLAUDE_CODE_SESSION_ID, else the newest)")
		await   = flag.Bool("await", false, "in -live, keep looking for a session instead of settling for the demo")
		mirror  = flag.Bool("mirror", true, "companion on the right, litter growing leftward (-mirror=false for the old left-anchored layout)")
		mockup  = flag.String("mockup", "", "write the left-vs-mirrored composition study to an HTML file")
		overlay = flag.String("overlay", "", "mock the agent INSIDE the scape: composite a captured pane (text file) over the scene")
		sandfd  = flag.String("sandfade", "", "tuner: how far the lower beach falls away to black")
		site    = flag.String("site", "", "write the submission page: reads <dir>/template.html, writes <dir>/index.html")
	)
	flag.Parse()

	// The half-block that smooths a gradient is Unicode, so ASCII-only output
	// cannot have it. Nothing else in the renderer is glyph-set dependent.
	if *asciiG {
		term.Shading = false
	}

	if *info {
		tw, th := termSize()
		fmt.Printf("profile=%s  size=%dx%d  glyph-chroma=%.1fx  sound=%s  TERM=%q COLORTERM=%q TERM_PROGRAM=%q\n",
			term.DetectProfile(), tw, th, term.GlyphBoost, notify.New().Describe(),
			os.Getenv("TERM"), os.Getenv("COLORTERM"), os.Getenv("TERM_PROGRAM"))
		return
	}

	act := scape.Activity{Working: *working, TimeOfDay: *tod, ContextUsed: *ctxUsed}
	if *todo != "" {
		if _, err := fmt.Sscanf(*todo, "%d/%d", &act.TodoDone, &act.TodoTotal); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes: -todo wants done/total, e.g. 3/5")
			os.Exit(2)
		}
	}
	switch {
	case *level >= 0:
		act.Level = *level
	case *working:
		act.Level = 0.7
	}

	c := canvas.New(*width, *height, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sc := scape.NewShore(*seed, *asciiG)
	profile := term.DetectProfile()

	if *live {
		wl, hl := 0, 0
		if isSet("w") {
			wl = *width
		}
		if isSet("h") {
			hl = *height
		}
		runLive(*seed, *fps, wl, hl, *ctxUsed, *tod, *asciiG, *session, *mirror, *await)
		return
	}

	if *sandfd != "" {
		if err := os.WriteFile(*sandfd, []byte(sandFadePage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		for _, ln := range sandFadeContrast(*seed, *tod) {
			fmt.Fprintln(os.Stderr, "  ", ln)
		}
		fmt.Println(*sandfd)
		return
	}

	if *overlay != "" {
		if flag.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "xscapes: -overlay needs a captured pane file, e.g. -overlay out.html pane.txt")
			os.Exit(2)
		}
		if err := os.WriteFile(*overlay, []byte(overlayPage(*seed, flag.Arg(0))), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*overlay)
		return
	}

	if *mockup != "" {
		if err := os.WriteFile(*mockup, []byte(mockupPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*mockup)
		return
	}

	if *facesHT != "" {
		if err := os.WriteFile(*facesHT, []byte(facePage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*facesHT)
		return
	}
	if *eyesHT != "" {
		if err := os.WriteFile(*eyesHT, []byte(eyesPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*eyesHT)
		return
	}

	if *colors != "" {
		if err := os.WriteFile(*colors, []byte(colorPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*colors)
		return
	}

	if *reel != "" {
		if err := os.WriteFile(*reel, []byte(reelPage(*seed, *reelAt, *reelN, *fps)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*reel)
		return
	}

	if *site != "" {
		page, err := sitePage(*seed, *site)
		if err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		out := filepath.Join(*site, "index.html")
		if err := os.WriteFile(out, []byte(page), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}
	if *wired != "" {
		if err := os.WriteFile(*wired, []byte(wiredPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*wired)
		return
	}

	if *kits != "" {
		if err := os.WriteFile(*kits, []byte(kittenPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*kits)
		return
	}

	if *busy != "" {
		if err := os.WriteFile(*busy, []byte(busyPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*busy)
		return
	}

	if *dayHTML != "" {
		if err := os.WriteFile(*dayHTML, []byte(dayPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*dayHTML)
		return
	}

	if *ctxdemo != "" {
		if err := os.WriteFile(*ctxdemo, []byte(contextPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*ctxdemo)
		return
	}

	if *layout != "" {
		if err := os.WriteFile(*layout, []byte(layoutPage(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*layout)
		return
	}

	if *strip != "" {
		if err := os.WriteFile(*strip, []byte(stripPage(*seed, *frames, *fps, *mode)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*strip)
		return
	}

	if *anim != "" {
		if err := os.WriteFile(*anim, []byte(animPage(*seed, 36, 12)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*anim)
		return
	}

	if *compare != "" {
		if err := os.WriteFile(*compare, []byte(compareSheet(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*compare)
		return
	}

	if *sheet != "" {
		if err := os.WriteFile(*sheet, []byte(contactSheet(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*sheet)
		return
	}

	if *html != "" {
		sc.Update(c, float64(*frames-1)/(*fps), act)
		if err := os.WriteFile(*html, []byte(c.RenderHTML("xscapes — "+sc.Name())), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*html)
		return
	}

	for i := 0; i < *frames; i++ {
		sc.Update(c, float64(i)/(*fps), act)
		if *plain {
			fmt.Println(c.RenderPlain())
		} else {
			fmt.Println(c.Render(profile))
		}
		if i < *frames-1 {
			fmt.Println()
		}
	}
}

// isSet reports whether a flag was given explicitly, so -live can size itself
// to the real terminal unless told otherwise.
func isSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// warnLegacyEnv names any pre-rename ASCIISCAPES_* variable still in the
// environment. They keep working -- see internal/envx -- but a setting whose
// name has moved is the easiest way there is to end up measuring nothing, so
// the program says which ones it saw rather than letting them pass.
//
// Here, at the top of main, because this is the last moment that is certainly
// safe to write to the terminal: `-live`, `claude` and `inside` take the
// alternate screen shortly after, and a stray line then lands in the middle of
// a painted frame. The two adapters are exempt -- `hook` runs on every tool
// call with its output already sent to /dev/null, and `statusline` owns a line
// of the agent's chrome that is not ours to scribble on.
func warnLegacyEnv(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "hook", "statusline":
			return
		}
	}
	envx.WarnLegacy(os.Stderr)
}
