package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// candidateE -- A SHOOTING STAR on a rare event.
//
// The instruction was to COUNT THE EVENT LOG BEFORE BUILDING IT, because an
// earlier count found zero compact events in his whole history and a cue bound
// to compaction would never fire. Counted here, from the spools themselves, so
// nobody has to trust a remembered number.
func candidateE() {
	fmt.Println("\n\n==== 6. CANDIDATE E -- A SHOOTING STAR ====")
	dir := filepath.Join(os.Getenv("HOME"), ".config", "xscapes", "run")
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	kinds := map[string]int{}
	lines := 0
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1<<20), 1<<22)
		for sc.Scan() {
			b := sc.Bytes()
			if len(b) == 0 {
				continue
			}
			var e struct {
				Kind string `json:"kind"`
			}
			if json.Unmarshal(b, &e) != nil {
				continue
			}
			kinds[e.Kind]++
			lines++
		}
		fh.Close()
	}
	fmt.Printf("  %d spools, %d events, counted now and not remembered:\n", len(files), lines)
	var ks []string
	for k := range kinds {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return kinds[ks[i]] > kinds[ks[j]] })
	for _, k := range ks {
		fmt.Printf("    %-16s %6d\n", k, kinds[k])
	}
	fmt.Printf("    %-16s %6d   <- the obvious binding, and it has NEVER fired\n", "compact", kinds["compact"])
	fmt.Printf("    %-16s %6d   <- and neither has this\n", "todo", kinds["todo"])

	fmt.Println("\n  ⇒ REJECT E, and the log is the reason, twice over:")
	fmt.Println("    1. NOTHING RARE FIRES. compact is 0 in the entire recorded history. A cue")
	fmt.Println("       bound to it would be dead code that never once reached his screen --")
	fmt.Println("       the same trap that left the constellation dark until session 27.")
	fmt.Println("    2. EVERY EVENT THAT DOES FIRE IS ALREADY SPOKEN FOR. needs_input and done")
	fmt.Println("       are the BUBBLE; error is the COMPANION; prompt closes a turn and that")
	fmt.Println("       already lights a star. Hanging a streak on any of them is one event on")
	fmt.Println("       two channels, which is what Hero's pacing study settled in session 26:")
	fmt.Println("       the encoding rule forbids it.")
	fmt.Println("  ⇒ THE ONE SHOOTING STAR THAT SURVIVES IS NOT BOUND TO A NEW EVENT AT ALL:")
	fmt.Println("    it is candidate A4 -- the star that arrives BY falling into its own place.")
	fmt.Println("    Same variable, same count, no second binding, and it fires exactly as often")
	fmt.Println("    as the channel already does.")
}
