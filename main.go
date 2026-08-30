// Command asciiscapes renders the thinking scene.
//
// Without flags it prints one frame, which is what you want in a pipe. The
// live TUI is a thin wrapper over exactly this renderer.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

func main() {
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
	)
	flag.Parse()

	if *info {
		fmt.Printf("profile=%s  TERM=%q COLORTERM=%q TERM_PROGRAM=%q\n",
			term.DetectProfile(), os.Getenv("TERM"), os.Getenv("COLORTERM"), os.Getenv("TERM_PROGRAM"))
		return
	}

	act := scape.Activity{Working: *working}
	switch {
	case *level >= 0:
		act.Level = *level
	case *working:
		act.Level = 0.7
	}

	c := canvas.New(*width, *height, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sc := scape.NewShore(*seed, *asciiG)
	profile := term.DetectProfile()

	if *sheet != "" {
		if err := os.WriteFile(*sheet, []byte(contactSheet(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "asciiscapes:", err)
			os.Exit(1)
		}
		fmt.Println(*sheet)
		return
	}

	if *html != "" {
		sc.Update(c, float64(*frames-1)/(*fps), act)
		if err := os.WriteFile(*html, []byte(c.RenderHTML("asciiscapes — "+sc.Name())), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "asciiscapes:", err)
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
