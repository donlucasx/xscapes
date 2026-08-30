package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// termSize asks stty rather than pulling in a dependency for one ioctl.
// Falls back to the design target when it cannot tell.
func termSize() (w, h int) {
	out, err := exec.Command("stty", "size").Output()
	if err == nil {
		if f := strings.Fields(string(out)); len(f) == 2 {
			if r, err1 := strconv.Atoi(f[0]); err1 == nil {
				if c, err2 := strconv.Atoi(f[1]); err2 == nil && r > 4 && c > 8 {
					return c, r
				}
			}
		}
	}
	return 80, 24
}

// runLive paints the scape to this terminal until interrupted. This is the
// only view that proves anything: the HTML harness models glyph advances, a
// terminal paints a fixed cell grid, and they disagree about non-ASCII glyphs.
func runLive(seed int64, fps float64, wIn, hIn int, ctxUsed float64, ascii bool) {
	w, h := wIn, hIn
	if w <= 0 || h <= 0 {
		w, h = termSize()
		h-- // leave the prompt a row to come back to
	}

	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, ascii)
	cat := companion.NewCat()
	_, chh := cat.Size()

	restore := func() { fmt.Print("\x1b[?25h\x1b[0m\n") }
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() { <-sig; restore(); os.Exit(0) }()

	fmt.Print("\x1b[?25l\x1b[2J") // hide cursor, clear once
	defer restore()

	profile := term.DetectProfile()
	tick := time.NewTicker(time.Duration(float64(time.Second) / fps))
	defer tick.Stop()

	start := time.Now()
	for range tick.C {
		t := time.Since(start).Seconds()
		// Swing through the states so every behaviour is visible in one sitting.
		st, act := companion.Working, scape.Activity{Working: true, Level: 0.65, ContextUsed: ctxUsed}
		switch phase := int(t/8) % 3; phase {
		case 1:
			st, act = companion.Resting, scape.Activity{ContextUsed: ctxUsed}
		case 2:
			st = companion.NeedsYou
		}

		sh.Update(c, t, act)
		top := c.H - 2 - chh
		cat.Draw(c.Near(), 5, top, t, st)
		if st == companion.NeedsYou {
			rows := companion.Bubble("tests passed")
			(&companion.Sprite{Rows: rows, Body: bubbleCol}).Draw(c.Near(), 12, top-len(rows))
		}

		// Home the cursor and overwrite rather than clearing: clearing every
		// frame is what makes a terminal animation flicker.
		fmt.Print("\x1b[H" + c.Render(profile))
	}
}
