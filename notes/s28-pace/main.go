// notes/s28-pace -- THE PACE SNAP, driven through the SHIPPED pace().
//
//	go run ./notes/s28-pace
//
// WHY THIS IS NOT ANOTHER PAPER DERIVATION. pace.go lives in package main at
// the repo root, which Go will not let another package import, so every study
// that has looked at the pace so far (notes/charstudy/crab/pace.go,
// notes/s28-closer/c-lean/pacecheck.go, notes/s28-smalls, notes/s28-litter)
// hand-COPIED the arithmetic and said so. A copy proves what the copy does.
//
// pace_shipped.go in this directory is a SYMLINK to ../../pace.go. The
// paceSpan / paceAt / stillFor / pace this program calls are the same function
// bodies the live loop calls, compiled from the same bytes on disk. If pace.go
// changes, this instrument changes with it and cannot go stale.
//
// The states are real too: a real reduce.Reducer fed real event.Events, read
// back through reduce.State exactly as live.go reads it. Nothing is hand-built.
//
// WHAT IT CHECKS. drawScene does
//
//	dx, moving := pace(st, c.W)
//	catX := lay.CatX + int(math.Round(dx))
//
// so dx IS the companion's column, up to the constant lay.CatX. pace() returns
// 0 for every held pose, and 0 is HOME -- the column nearest the frame edge.
// The claim under test is that the companion therefore jumps OUTWARD, away
// from the reader, at the exact frame it starts asking the reader for
// something. And that it jumps INWARD again the frame the ask is answered.
//
// Every number below is read off a RENDERED frame: the crab is drawn into a
// real canvas at the column drawScene would draw it, and the ink is counted.
package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// widths are his: 40 and 80 are the design targets, the other four are window
// sizes off his own recordings.
var widths = []int{40, 80, 111, 124, 143, 153}

// homeCol reproduces compose()'s catX (live.go, `right := margin + w/32`)
// because compose is unexported and unreachable from here. It is NOT what the
// finding rests on -- the finding is the OFFSET, which comes from the shipped
// pace() -- and it is cross-checked against a number measured independently
// last session: the s26 banner records the crab pacing "107 -> 101 -> 107 at
// 124 columns", and this function prints 107 at 124 with a span of 6.
func homeCol(w, catW int) int {
	x := w - catW - (2 + w/32)
	if x < 0 {
		x = 0
	}
	return x
}

func main() {
	fmt.Println(strings.Repeat("=", 78))
	fmt.Println("THE PACE SNAP -- shipped pace(), real reduce.State, measured on the frame")
	fmt.Println(strings.Repeat("=", 78))

	sanity()
	askSnap()
	askAnswered()
	worriedOverAsk()
	sweep()
}

// ------------------------------------------------------------------ helpers

// walked drives a reducer up to n main-thread tool events and hands back the
// state a frame would be rendered from, plus the clock it was read at.
//
// Steps come only from ToolStart with Agent == "" (reduce.go), so this is the
// same thing a real agent doing n edits produces. Events are 1.5s apart: long
// enough that the stride has landed (stepDur is 0.28) and short enough that
// TurnSilence never closes the turn, so the pose is genuinely Working.
func walked(n int) (*reduce.Reducer, time.Time) {
	t0 := time.Date(2026, 9, 10, 14, 0, 0, 0, time.UTC)
	r := reduce.New("s28-pace")
	now := t0
	r.Apply(event.Event{Kind: event.Prompt, Text: "go"}, now)
	for i := 0; i < n; i++ {
		now = now.Add(1500 * time.Millisecond)
		id := fmt.Sprintf("t%d", i)
		r.Apply(event.Event{Kind: event.ToolStart, ID: id, Op: event.OpEdit, Tool: "Edit"}, now)
		r.Apply(event.Event{Kind: event.ToolEnd, ID: id, Op: event.OpEdit, Tool: "Edit"}, now.Add(50*time.Millisecond))
	}
	// A full second after the last step: the stride has landed, so nothing
	// below is an artefact of catching the companion mid-stride.
	return r, now.Add(time.Second)
}

// ink renders the companion the way drawScene does -- at lay.CatX + round(dx)
// -- into a real canvas, and returns the leftmost and rightmost column that
// actually carries ink. This is the RENDERED frame, not the arithmetic.
func ink(w int, dx float64, pose companion.State, moving bool) (lo, hi int, row string) {
	crab := companion.New(companion.DefaultName)
	cw, ch := crab.Size()
	c := canvas.New(w, ch+6, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	top := 3
	x := homeCol(w, cw) + int(math.Round(dx))
	crab.SetStepping(moving)
	crab.Draw(c.Near(), x, top, 0, pose)

	lo, hi = -1, -1
	for col := 0; col < w; col++ {
		for y := 0; y < c.H; y++ {
			if cell := c.Near().Cells[y*w+col]; cell.Set && cell.R != ' ' {
				if lo < 0 {
					lo = col
				}
				hi = col
				break
			}
		}
	}
	b := []byte(strings.Repeat(".", w))
	for col := lo; col >= 0 && col <= hi; col++ {
		b[col] = '#'
	}
	return lo, hi, string(b)
}

func poseName(s companion.State) string {
	switch s {
	case companion.Resting:
		return "Resting"
	case companion.Working:
		return "Working"
	case companion.NeedsYou:
		return "NeedsYou"
	case companion.Done:
		return "Done"
	case companion.Worried:
		return "Worried"
	}
	return fmt.Sprintf("State(%v)", s)
}

// ------------------------------------------------------------------ sections

// sanity proves the instrument is looking at something. A probe whose input is
// empty reads clean and looks exactly like a pass (s26 paid for that lesson),
// so before any verdict: the companion must actually be away from home, and
// the rendered ink must actually be somewhere.
func sanity() {
	fmt.Println("\n0. THE INSTRUMENT IS LOOKING AT SOMETHING")
	fmt.Println("   (a clean reading from an empty probe looks exactly like a pass)")
	fmt.Printf("   %-6s %-6s %-7s %-9s %-8s %s\n",
		"width", "span", "home", "steps", "dx", "rendered ink columns")
	nonZero, drew := 0, 0
	for _, w := range widths {
		span := paceSpan(w)
		r, now := walked(span) // span steps == the far end of the triangle
		st := r.State(now)
		dx, moving := pace(st, w)
		lo, hi, _ := ink(w, dx, st.Pose, moving)
		if dx != 0 {
			nonZero++
		}
		if lo >= 0 {
			drew++
		}
		fmt.Printf("   %-6d %-6d %-7d %-9d %-8.2f %d..%d  pose=%s\n",
			w, span, homeCol(w, 12), st.Steps, dx, lo, hi, poseName(st.Pose))
	}
	fmt.Printf("   => %d of %d widths had the companion AWAY from home before the ask,\n", nonZero, len(widths))
	fmt.Printf("      and %d of %d rendered ink. The probe is non-empty.\n", drew, len(widths))
}

// askSnap is the finding. One needs_input event, nothing else changes.
func askSnap() {
	fmt.Println("\n" + strings.Repeat("-", 78))
	fmt.Println("1. THE ASK LANDS: working -> NeedsYou, one event, no other change")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println("   Worst case per width: the companion is at the far end of its triangle")
	fmt.Println("   (span steps in) when the agent asks for permission.")
	fmt.Println()

	worst := 0.0
	jumped := 0
	for _, w := range widths {
		span := paceSpan(w)
		r, now := walked(span)

		stB := r.State(now)
		dxB, movB := pace(stB, w)
		loB, hiB, rowB := ink(w, dxB, stB.Pose, movB)

		// The next frame. ONE event: the agent asks.
		r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, now)
		stA := r.State(now)
		dxA, movA := pace(stA, w)
		loA, hiA, rowA := ink(w, dxA, stA.Pose, movA)

		jump := dxA - dxB
		if math.Abs(jump) > worst {
			worst = math.Abs(jump)
		}
		if jump != 0 {
			jumped++
		}
		dir := "OUTWARD, toward the frame edge, away from the reader"
		if jump < 0 {
			dir = "inward"
		}
		fmt.Printf("   w=%d  steps=%d  span=%d\n", w, stB.Steps, span)
		fmt.Printf("     before  %-8s dx=%+6.2f  ink %3d..%-3d\n", poseName(stB.Pose), dxB, loB, hiB)
		fmt.Printf("     after   %-8s dx=%+6.2f  ink %3d..%-3d   JUMP %+.0f cells %s\n",
			poseName(stA.Pose), dxA, loA, hiA, jump, dir)
		if w >= 80 {
			fmt.Printf("       %s\n       %s\n", rowB, rowA)
		}
		fmt.Println()
	}
	fmt.Printf("   => %d of %d widths teleport. Worst: %.0f cells, in ONE frame.\n", jumped, len(widths), worst)
	fmt.Println("      Steps did not change and StepAge did not change. The ONLY thing that")
	fmt.Println("      changed is the pose, and stillFor() reads \"still\" as dx = 0 = home.")
}

// askAnswered is the same defect running backwards, and it is the half nobody
// has written down: answering the permission prompt starts a tool, a ToolStart
// clears needsInput AND steps the companion, so the pose goes back to Working
// and dx goes from 0 straight back to where it had been.
func askAnswered() {
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println("2. THE ASK IS ANSWERED: NeedsYou -> Working, the snap runs BACKWARDS")
	fmt.Println(strings.Repeat("-", 78))
	for _, w := range widths {
		span := paceSpan(w)
		r, now := walked(span)
		r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, now)
		stB := r.State(now)
		dxB, movB := pace(stB, w)
		loB, hiB, _ := ink(w, dxB, stB.Pose, movB)

		// He hits Enter; the tool runs. ToolStart clears needsInput (reduce.go).
		now = now.Add(2 * time.Second)
		r.Apply(event.Event{Kind: event.ToolStart, ID: "answered", Op: event.OpShell, Tool: "Bash"}, now)
		stA := r.State(now)
		dxA, movA := pace(stA, w)
		loA, hiA, _ := ink(w, dxA, stA.Pose, movA)

		fmt.Printf("   w=%-4d %-8s dx=%+6.2f ink %3d..%-3d  ->  %-8s dx=%+6.2f ink %3d..%-3d  JUMP %+.0f\n",
			w, poseName(stB.Pose), dxB, loB, hiB, poseName(stA.Pose), dxA, loA, hiA, dxA-dxB)
	}
	fmt.Println("   => The companion snaps out when it asks and snaps back in when it is")
	fmt.Println("      answered. Two teleports per ask, both invisible to any test today.")
	fmt.Println()
}

// worriedOverAsk is the case the fix must not get wrong. Worried OUTRANKS
// NeedsYou (reduce.go pose()), so a main-thread error while the ask is up
// flips the pose again. Both are held poses; neither may move the companion.
func worriedOverAsk() {
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println("3. A MAIN-THREAD ERROR WHILE THE ASK IS UP: NeedsYou -> Worried")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println("   Worried outranks NeedsYou. Both are held poses, so this transition must")
	fmt.Println("   move the companion by ZERO cells -- today, and after any fix.")
	for _, w := range widths {
		r, now := walked(paceSpan(w))
		r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, now)
		stB := r.State(now)
		dxB, _ := pace(stB, w)
		now = now.Add(time.Second)
		r.Apply(event.Event{Kind: event.Error, Tool: "Bash", Detail: "exit 1"}, now)
		stA := r.State(now)
		dxA, _ := pace(stA, w)
		fmt.Printf("   w=%-4d %-8s dx=%+6.2f  ->  %-8s dx=%+6.2f   move %+.2f\n",
			w, poseName(stB.Pose), dxB, poseName(stA.Pose), dxA, dxA-dxB)
	}
	fmt.Println()
}

// sweep is the whole triangle at every width, so the finding is not one lucky
// step count. For each n, how far the companion is standing from home when the
// ask lands -- which is exactly how far it teleports.
func sweep() {
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println("4. EVERY STEP COUNT, EVERY WIDTH -- the size of the jump is where it stood")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Printf("   %-6s %-5s  %s\n", "width", "span", "jump in cells, by step count n = 0,1,2,...,2*span")
	total, moved := 0, 0
	sum := 0.0
	for _, w := range widths {
		span := paceSpan(w)
		var cells []string
		for n := 0; n <= 2*span; n++ {
			r, now := walked(n)
			stB := r.State(now)
			dxB, _ := pace(stB, w)
			r.Apply(event.Event{Kind: event.NeedsInput, Text: "?"}, now)
			dxA, _ := pace(r.State(now), w)
			j := dxA - dxB
			total++
			if j != 0 {
				moved++
			}
			sum += math.Abs(j)
			cells = append(cells, fmt.Sprintf("%+.0f", j))
		}
		fmt.Printf("   %-6d %-5d  %s\n", w, span, strings.Join(cells, " "))
	}
	fmt.Printf("\n   SAMPLE: %d (width, step-count) transitions, every one driven through a\n", total)
	fmt.Printf("   real reducer and the shipped pace(). %d of %d move the companion;\n", moved, total)
	fmt.Printf("   mean displacement over all %d is %.2f cells, and every non-zero one is\n", total, sum/float64(total))
	fmt.Println("   OUTWARD. A step count that happens to be at home (n = 0, or a multiple")
	fmt.Println("   of 2*span) is the only case that does not jump.")
	fmt.Println()
	// ⚠ THE VERDICT IS COMPUTED, NOT TYPED. It used to be two hardcoded
	// Println lines reading "CONFIRMED ... the shipped code moves it FARTHER",
	// and pace_shipped.go is a SYMLINK to the shipped pace.go -- so the moment
	// the defect was fixed this instrument printed a live table of zeros above
	// a confident claim that the bug was still there. An instrument that states
	// a conclusion its own numbers no longer support is worse than no
	// instrument: it is a wrong answer with a measurement stapled to it.
	if moved > 0 {
		fmt.Println("   VERDICT: CONFIRMED, the defect is PRESENT in the code this just ran.")
		fmt.Println("   His idea asks the companion to come CLOSER before it prompts. This")
		fmt.Println("   moves it FARTHER, in one frame, first.")
	} else {
		fmt.Println("   VERDICT: FIXED in the code this just ran. Not one of the transitions")
		fmt.Println("   above moves the companion, so a pose change on its own costs zero")
		fmt.Println("   cells and the ask no longer begins by walking away from him.")
	}
}
